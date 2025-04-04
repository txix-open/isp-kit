package grpc_test

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
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
	"net"
	"net/http"
	"testing"
	"time"
)

type Data struct {
	Value string
}

var (
	handler = func(ctx context.Context, req Data) *Data {
		return &Data{Value: req.Value}
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

func BenchmarkHttp11Parallel(b *testing.B) {
	test, _ := test.New(&testing.T{})
	srv := httpt.NewMock(test)
	srv.Wrapper = endpoint.DefaultWrapper(test.Logger(), httplog.Noop())
	srv.POST("/echo", handler)
	cli := srv.Client(httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...))
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

func BenchmarkHttp2H2CParallel(b *testing.B) {
	test, _ := test.New(&testing.T{})
	h2s := &http2.Server{}
	wrapper := endpoint.DefaultWrapper(test.Logger(), httplog.Noop())
	router := router2.New().POST("/echo", wrapper.Endpoint(handler))
	server := &http.Server{
		Handler: h2c.NewHandler(router, h2s),
	}
	err := http2.ConfigureServer(server, h2s)
	if err != nil {
		b.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		b.Fatal(err)
	}
	go server.Serve(listener)
	time.Sleep(100 * time.Millisecond)

	client := http.Client{
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
	cli := httpcli.NewWithClient(&client, httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...))
	cli.GlobalRequestConfig().BaseUrl = "http://" + listener.Addr().String()
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

func BenchmarkHttp2Parallel(b *testing.B) {
	test, _ := test.New(&testing.T{})
	wrapper := endpoint.DefaultWrapper(test.Logger(), httplog.Noop())
	router := router2.New().POST("/echo", wrapper.Endpoint(handler))
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go server.ListenAndServeTLS("_certificates/server.crt", "_certificates/server.key")
	time.Sleep(100 * time.Millisecond)

	transport := httpcli.StdClient.Transport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := http.Client{
		Transport: transport,
	}
	cli := httpcli.NewWithClient(&client, httpcli.WithMiddlewares(httpclix.DefaultMiddlewares()...))
	cli.GlobalRequestConfig().BaseUrl = "https://localhost:8080"
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
