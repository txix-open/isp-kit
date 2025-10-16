package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	endpoint2 "github.com/txix-open/isp-kit/http/endpoint/v2"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/router"
	"github.com/txix-open/isp-kit/test"
)

type MockServer struct {
	Wrapper endpoint2.Wrapper
	srv     *httptest.Server
	router  *router.Router
}

func NewMock(t *test.Test) *MockServer {
	router := router.New()
	srv := httptest.NewServer(router)
	t.T().Cleanup(func() {
		srv.Close()
	})
	wrapper := endpoint.DefaultWrapper(t.Logger(), httplog.Log(t.Logger(), true))
	return &MockServer{
		Wrapper: wrapper,
		srv:     srv,
		router:  router,
	}
}

func (m *MockServer) Client(opts ...httpcli.Option) *httpcli.Client {
	cli := httpcli.NewWithClient(m.srv.Client(), opts...)
	cli.GlobalRequestConfig().BaseUrl = m.BaseURL()
	return cli
}

func (m *MockServer) BaseURL() string {
	return m.srv.URL
}

func (m *MockServer) POST(path string, handler endpoint2.Wrappable) *MockServer {
	return m.Mock(http.MethodPost, path, handler)
}

func (m *MockServer) GET(path string, handler endpoint2.Wrappable) *MockServer {
	return m.Mock(http.MethodGet, path, handler)
}

func (m *MockServer) Mock(method string, path string, handler endpoint2.Wrappable) *MockServer {
	m.router.Handler(method, path, handler.Wrap(m.Wrapper))
	return m
}
