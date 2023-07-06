package httpclix_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/httpclix"
	"github.com/integration-system/isp-kit/http/httperrors"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/httpt"
	"github.com/pkg/errors"
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
			return nil, httperrors.New(http.StatusBadRequest, errors.New("test error"))
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
