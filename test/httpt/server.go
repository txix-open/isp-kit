package httpt

import (
	"net/http"
	"net/http/httptest"

	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/http/httpclix"
	"gitlab.txix.ru/isp/isp-kit/test"
)

func TestServer(t *test.Test, handler http.Handler, opts ...httpcli.Option) (*httptest.Server, *httpcli.Client) {
	srv := httptest.NewServer(handler)
	cli := httpclix.Default(opts...)
	cli.GlobalRequestConfig().BaseUrl = srv.URL
	return srv, cli
}
