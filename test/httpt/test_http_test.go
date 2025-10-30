package httpt_test

import (
	"context"
	"net/http"
	"testing"

	endpoint2 "github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/httpt"
)

type resp struct {
	Ok bool
}

func Test(t *testing.T) {
	t.Parallel()
	test, assert := test.New(t)
	mock := httpt.NewMock(test)
	mock.POST("/endpoint1", endpoint2.NewDefaultHttp(func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return json.NewEncoder(w).Encode(resp{Ok: true})
	}))
	mock.GET("/endpoint2", endpoint2.NewDefaultHttp(func(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
		return json.NewEncoder(w).Encode(resp{Ok: false})
	}))

	client := mock.Client()

	resp1 := resp{}
	r, err := client.Post("/endpoint1").JsonResponseBody(&resp1).Do(t.Context())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: true}, resp1)

	resp2 := resp{}
	r, err = client.Get("/endpoint2").JsonResponseBody(&resp2).Do(t.Context())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
	assert.EqualValues(resp{Ok: false}, resp2)

	r, err = client.Get("endpoint2").Do(t.Context())
	assert.NoError(err)
	assert.EqualValues(http.StatusOK, r.StatusCode())
}
