package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/test"
)

// TestServer creates and starts an HTTP test server with the provided handler,
// returning both the server and a configured client. The server listens on
// a random local port and the client is pre-configured with the server's URL.
func TestServer(t *test.Test, handler http.Handler, opts ...httpcli.Option) (*httptest.Server, *httpcli.Client) {
	srv := httptest.NewServer(handler)
	cli := httpclix.Default(opts...)
	cli.GlobalRequestConfig().BaseUrl = srv.URL
	return srv, cli
}
