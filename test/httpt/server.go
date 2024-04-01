package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/test"
)

func TestServer(t *test.Test, handler http.Handler, opts ...httpcli.Option) (*httptest.Server, *httpcli.Client) {
	srv := httptest.NewServer(handler)
	cli := httpclix.Default(opts...)
	cli.GlobalRequestConfig().BaseUrl = srv.URL
	return srv, cli
}
