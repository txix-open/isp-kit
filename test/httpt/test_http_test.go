package httpt_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/integration-system/isp-kit/test"
	"github.com/integration-system/isp-kit/test/httpt"
)

type resp struct {
	Ok bool
}

func Test(t *testing.T) {
	test, assert := test.New(t)
	mock := httpt.NewMock(test)
	mock.POST("/endpoint1", func() resp {
		return resp{Ok: true}
	}).GET("/endpoint2", func(ctx context.Context) resp {
		return resp{Ok: false}
	})

	resty := resty.NewWithClient(mock.Client()).SetBaseURL(mock.BaseURL())

	resp1 := resp{}
	r, err := resty.R().SetResult(&resp1).Post("/endpoint1")
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: true}, resp1)

	resp2 := resp{}
	r, err = resty.R().SetResult(&resp2).Get("/endpoint2")
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: false}, resp2)
}
