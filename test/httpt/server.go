package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/httpclix"
	"github.com/integration-system/isp-kit/test"
)

func TestServer(t *test.Test, handler http.Handler, opts ...httpcli.Option) (*httptest.Server, *httpcli.Client) {
	srv := httptest.NewServer(handler)
	opts = append(opts, httpcli.WithGlobalRequestConfig(httpcli.GlobalRequestConfig{
		BaseUrl: srv.URL,
	}))
	cli := httpclix.Default(opts...)
	return srv, cli
}
