package http_test

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	isphttp "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

type Request struct {
	Id string `validate:"required"`
}

type Response struct {
	Result string
}

func TestService(t *testing.T) {
	t.Parallel()

	url := prepareServer(t)
	response := Response{}

	client := resty.New().SetBaseURL("http://" + url)
	resp, err := client.R().
		SetBody(Request{Id: "man"}).
		SetResult(&response).
		Post("/getId")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected := Response{Result: "Hello_man"}
	require.Equal(t, expected, response)

	resp, err = client.R().
		SetBody(Request{Id: ""}).
		Post("/getId")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

	resp, err = client.R().
		SetBody(Request{Id: "smth"}).
		Post("/badGetId")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode())

	response = Response{}
	resp, err = client.R().
		SetResult(&response).
		Get("/noBody")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected = Response{Result: "Test"}
	require.Equal(t, expected, response)
}

func TestRecover(t *testing.T) {
	t.Parallel()
	url := prepareServer(t)

	response := Response{}
	client := resty.New().SetBaseURL("http://" + url)

	resp, err := client.R().
		SetBody(Request{Id: "man"}).
		SetResult(&response).
		Post("/recover")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode())

	response = Response{}
	resp, err = client.R().
		SetResult(&response).
		Get("/noBody")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected := Response{Result: "Test"}
	require.Equal(t, expected, response)
}

type endpointDescriptor struct {
	Path    string
	Handler endpoint.Wrappable
}

func prepareServer(t *testing.T) string {
	t.Helper()

	logger, err := log.New(log.WithLevel(log.DebugLevel))
	require.NoError(t, err)

	endpoints := []endpointDescriptor{{
		Path: "/getId",
		Handler: endpoint.New(func(ctx context.Context, req Request) (*Response, error) {
			return &Response{Result: "Hello_" + req.Id}, nil
		}),
	}, {
		Path: "/badGetId",
		Handler: endpoint.New(func(ctx context.Context, req Request) (*Response, error) {
			return &Response{}, apierrors.New(http.StatusNotFound, http.StatusNotFound, "not found", errors.New("Not Found"))
		}),
	}, {
		Path: "/noBody",
		Handler: endpoint.NewDefaultHttp(func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
			w.Header().Set("content-type", "application/json")
			return json.EncodeInto(w, Response{Result: "Test"})
		}),
	}, {
		Path: "/recover",
		Handler: endpoint.NewWithRequest(func(ctx context.Context, _ *http.Request) error {
			panic(errors.New("test panic error"))
		}),
	}}

	mapper := endpoint.DefaultWrapper(logger, httplog.Log(logger, true))
	muxer := http.NewServeMux()
	for _, descriptor := range endpoints {
		muxer.Handle(descriptor.Path, mapper.EndpointV2(descriptor.Handler))
	}

	var lc net.ListenConfig
	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:")
	require.NoError(t, err)

	srv := isphttp.NewServer(logger)
	srv.Upgrade(muxer)
	go func() {
		err := srv.Serve(listener)
		assert.NoError(t, err)
	}()

	return listener.Addr().String()
}
