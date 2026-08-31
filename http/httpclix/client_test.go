package httpclix_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/txix-open/isp-kit/codec"
	"github.com/txix-open/isp-kit/errors"
	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/httpt"
)

type example struct {
	Data string
}

func TestDefault(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	srv := httpt.NewMock(test)
	invokeNumber := 0
	url := srv.POST("/api/save", endpoint.New(func(ctx context.Context, req example) (*example, error) {
		require.EqualValues(expectedId, requestid.FromContext(ctx))

		invokeNumber++
		if invokeNumber == 1 {
			return nil, apierrors.NewBusinessError(http.StatusBadRequest, "test error", errors.New("test error"))
		}
		return &req, nil
	})).BaseURL()

	cli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(test.Logger())))
	exp := example{}
	ctx = http_metrics.ClientEndpointToContext(ctx, "/api/save")
	resp, err := cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.EqualValues(http.StatusBadRequest, resp.StatusCode())

	exp = example{}
	resp, err = cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestLogHeaders(t *testing.T) {
	t.Parallel()
	testEnv, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	srv := httpt.NewMock(testEnv)
	url := srv.POST("/api/save", endpoint.New(func(ctx context.Context, req example) (*example, error) {
		return &req, nil
	})).BaseURL()

	cli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(testEnv.Logger())))

	ctx = httpclix.LogConfigToContext(ctx, false, false,
		httpclix.LogHeaders(true, true),
		httpclix.LogDump(true, true),
		httpclix.LogCombined(true),
	)

	exp := example{}

	resp, err := cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestLog_Encoded(t *testing.T) {
	t.Parallel()
	testEnv, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	codec := codec.Default
	srv := httpt.NewMock(testEnv)
	srv.Wrapper.Middlewares = []http2.Middleware{
		endpoint.RequestId(),
		endpoint.DecodeRequest(codec),
		endpoint.EncodeResponse(codec, 1),
		http2.Middleware(httplog.CombinedLog(testEnv.Logger(), true)),
		endpoint.ErrorHandler(testEnv.Logger()),
		endpoint.Recovery(),
	}
	url := srv.POST("/api/save", func(ctx context.Context, req example) (*example, error) {
		return &req, nil
	}).BaseURL()

	cli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(testEnv.Logger())))

	ctx = httpclix.LogConfigToContext(ctx, true, true,
		httpclix.LogCombined(true),
	)

	exp := example{}

	resp, err := cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		AcceptEncodedResponse(true).
		EncodeRequest(true).
		EncodeThreshold(5).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.IsSuccess())
}

func TestLogBodies(t *testing.T) {
	t.Parallel()
	testEnv, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(t.Context(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	srv := httpt.NewMock(testEnv)
	url := srv.POST("/api/save", endpoint.New(func(ctx context.Context, req example) (*example, error) {
		return &req, nil
	})).BaseURL()

	cli := httpclix.Default(httpcli.WithMiddlewares(httpclix.LogWithOptions(testEnv.Logger(),
		httpclix.LogBody(true, true),
		httpclix.LogCombined(true),
	)))

	exp := example{}

	resp, err := cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.IsSuccess())
}
