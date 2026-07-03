package http_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/codec"
	isphttp "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/test/fake"
)

type Request struct {
	Id string `validate:"required"`
}

type Response struct {
	Result string
}

//nolint:funlen
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

	encodedReq, err := json.Marshal(Request{Id: "man_1"})
	require.NoError(t, err)

	codec := codec.Default

	encodedReq, err = codec.EncodeBytes(encodedReq)
	require.NoError(t, err)

	resp, err = client.R().
		SetBody(encodedReq).
		SetHeader("Content-Encoding", codec.Type()).
		SetResult(&response).
		Post("/getIdEncoded")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected = Response{Result: "Hello_man_1"}
	require.Equal(t, expected, response)

	resp, err = client.R().
		SetBody(Request{Id: "man"}).
		SetHeader("Accept-Encoding", codec.Type()).
		SetResult(&response).
		Post("/getIdEncoded")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected = Response{Result: "Hello_man"}
	require.Equal(t, expected, response)

	bigRequest := Request{
		Id: string(fake.It[[]byte](fake.MinSliceSize(1024), fake.MaxSliceSize(2*1024))),
	}

	resp, err = client.R().
		SetBody(bigRequest).
		SetHeader("Accept-Encoding", codec.Type()).
		Post("/getIdEncoded")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

	expected = Response{Result: fmt.Sprintf("Hello_%s", bigRequest.Id)}

	respDeEncoded, err := codec.DecodeBytes(resp.Body())
	require.NoError(t, err)

	err = json.Unmarshal(respDeEncoded, &response)
	require.NoError(t, err)

	require.Equal(t, expected, response)

	resp, err = client.R().
		SetBody(bigRequest).
		SetResult(&response).
		Post("/getIdEncoded")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())

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

	endpoints = []endpointDescriptor{
		{
			Path: "/getIdEncoded",
			Handler: endpoint.New(func(ctx context.Context, req Request) (*Response, error) {
				return &Response{Result: "Hello_" + req.Id}, nil
			}),
		},
	}

	codecMapper := endpoint.DefaultWrapper(
		logger,
		httplog.Log(logger, true),
	)
	codec := codec.Default
	codecMapper.Middlewares = []isphttp.Middleware{
		endpoint.MaxRequestBodySize(64 * 1024 * 1024),
		endpoint.RequestId(),
		endpoint.EncodeResponse(codec, 1024),
		endpoint.DecodeRequest(codec),
		isphttp.Middleware(httplog.Log(logger, true)),
		endpoint.ErrorHandler(logger),
		endpoint.Recovery(),
	}
	for _, descriptor := range endpoints {
		muxer.Handle(descriptor.Path, codecMapper.EndpointV2(descriptor.Handler))
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
