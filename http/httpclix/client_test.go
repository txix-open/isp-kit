package httpclix_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
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
	test, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(context.Background(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	invokeNumber := 0
	url := httpt.NewMock(test).POST("/api/save", func(ctx context.Context, req example) (*example, error) {
		require.EqualValues(expectedId, requestid.FromContext(ctx))

		invokeNumber++
		if invokeNumber == 1 {
			return nil, apierrors.NewBusinessError(http.StatusBadRequest, "test error", errors.New("test error"))
		}
		return &req, nil
	}).BaseURL()

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
	testEnv, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(context.Background(), expectedId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedId))

	url := httpt.NewMock(testEnv).POST("/api/save", func(ctx context.Context, req example) (*example, error) {
		return &req, nil
	}).BaseURL()

	cli := httpclix.Default(httpcli.WithMiddlewares(httpclix.Log(testEnv.Logger())))

	ctx = httpclix.LogConfigToContext(ctx, false, false,
		httpclix.LogHeaders(true, true),
		httpclix.LogDump(true, true),
	)

	exp := example{}

	resp, err := cli.Post(url + "/api/save").
		JsonRequestBody(example{"test"}).
		JsonResponseBody(&exp).
		Do(ctx)
	require.NoError(err)
	require.True(resp.IsSuccess())
}
