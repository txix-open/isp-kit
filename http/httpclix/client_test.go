package httpclix_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/http/apierrors"
	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/http/httpclix"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics/http_metrics"
	"gitlab.txix.ru/isp/isp-kit/requestid"
	"gitlab.txix.ru/isp/isp-kit/test"
	"gitlab.txix.ru/isp/isp-kit/test/httpt"
)

type example struct {
	Data string
}

func TestDefault(t *testing.T) {
	test, require := test.New(t)

	expectedId := requestid.Next()
	ctx := requestid.ToContext(context.Background(), expectedId)
	ctx = log.ToContext(ctx, log.String("requestId", expectedId))

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
