package httpt_test

import (
	"context"
	"net/http"
	"testing"

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

	client := mock.Client()

	resp1 := resp{}
	r, err := client.Post("/endpoint1").JsonResponseBody(&resp1).Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: true}, resp1)

	resp2 := resp{}
	r, err = client.Get("/endpoint2").JsonResponseBody(&resp2).Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: false}, resp2)

	r, err = client.Get("endpoint2").Do(context.Background())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
}
