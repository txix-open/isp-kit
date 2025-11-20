// nolint:forcetypeassert,gosec
package grpc_test

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	router2 "github.com/txix-open/isp-kit/http/router"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grpct"
	"github.com/txix-open/isp-kit/test/httpt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Data struct {
	Value string
}

// nolint:gochecknoglobals
var (
	handler = func(ctx context.Context, req Data) (*Data, error) {
		return &Data{Value: req.Value}, nil
	}
)

func BenchmarkGrpcParallel(b *testing.B) {
	test, _ := test.New(&testing.T{})
	srv, cli := grpct.NewMock(test)
	srv.Mock("echo", handler)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			resp := Data{}
			_ = cli.Invoke("echo").
				JsonRequestBody(Data{Value: hex.EncodeToString(data)}).
				JsonResponseBody(&resp).
				Do(b.Context())
		}
	})
}

func BenchmarkHttp11Parallel_V2(b *testing.B)    { runHTTPBenchmark(b, true, false, false) }
func BenchmarkHttp11Parallel_Ref(b *testing.B)   { runHTTPBenchmark(b, false, false, false) }
func BenchmarkHttp2H2CParallel_V2(b *testing.B)  { runHTTPBenchmark(b, true, true, false) }
func BenchmarkHttp2H2CParallel_Ref(b *testing.B) { runHTTPBenchmark(b, false, true, false) }
func BenchmarkHttp2Parallel_V2(b *testing.B)     { runHTTPBenchmark(b, true, true, true) }
func BenchmarkHttp2Parallel_Ref(b *testing.B)    { runHTTPBenchmark(b, false, true, true) }

// nolint:thelper
func runHTTPBenchmark(b *testing.B, useV2, h2, tlsEnabled bool) {
	test, _ := test.New(&testing.T{})
	wrapper := endpoint.DefaultWrapper(test.Logger(), httplog.Noop())

	cli, baseURL := newHTTPClient(b, h2, useV2, tlsEnabled, wrapper, test)

	cli.GlobalRequestConfig().BaseUrl = baseURL
	cli.GlobalRequestConfig().Timeout = 15 * time.Second

	b.ResetTimer()
	b.ReportAllocs()
	b.SetParallelism(32)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			data := make([]byte, 4200)
			_, _ = rand.Read(data)
			resp := Data{}
			_ = cli.Post("/echo").
				JsonRequestBody(Data{Value: hex.EncodeToString(data)}).
				JsonResponseBody(&resp).
				DoWithoutResponse(b.Context())
		}
	})
}

func newHTTPClient(b *testing.B, h2, useV2, tlsEnabled bool, wrapper endpoint.Wrapper, test *test.Test) (*httpcli.Client, string) {
	b.Helper()

	if !h2 {
		// HTTP/1.1 Mock server
		srv := httpt.NewMock(test)
		srv.Wrapper = wrapper
		if useV2 {
			srv.POST("/echo", wrapper.EndpointV2(endpoint.New(handler)))
		} else {
			srv.POST("/echo", wrapper.Endpoint(handler))
		}
		return srv.Client(httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...)), srv.BaseURL()
	}

	// HTTP/2 branch
	h2s := &http2.Server{}
	var router http.Handler
	if useV2 {
		router = router2.New().POST("/echo", wrapper.EndpointV2(endpoint.New(handler)))
	} else {
		router = router2.New().POST("/echo", wrapper.Endpoint(handler))
	}

	server := &http.Server{}
	if tlsEnabled {
		server.Addr = ":8443"
		server.Handler = router
		go func() {
			_ = server.ListenAndServeTLS("_certificates/server.crt", "_certificates/server.key")
		}()

		transport := httpcli.StdClient.Transport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		cli := httpcli.NewWithClient(
			&http.Client{Transport: transport},
			httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...),
		)
		return cli, "https://localhost:8443"
	}

	// H2C (no TLS)
	var lc net.ListenConfig
	listener, _ := lc.Listen(b.Context(), "tcp", "127.0.0.1:0")
	server.Handler = h2c.NewHandler(router, h2s)
	go func() {
		_ = server.Serve(listener)
	}()

	cli := &http.Client{
		Transport: &http2.Transport{
			// So http2.Transport doesn't complain the URL scheme isn't 'https'
			AllowHTTP: true,
			// Pretend we are dialing a TLS endpoint. (Note, we ignore the passed tls.Config)
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
	}
	return httpcli.NewWithClient(cli, httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...)), "http://" + listener.Addr().String()
}
