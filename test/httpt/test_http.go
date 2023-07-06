package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/integration-system/isp-kit/http/endpoint"
	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/router"
	"github.com/integration-system/isp-kit/test"
)

type MockServer struct {
	wrapper endpoint.Wrapper
	srv     *httptest.Server
	router  *router.Router
}

func NewMock(t *test.Test) *MockServer {
	router := router.New()
	srv := httptest.NewServer(router)
	t.T().Cleanup(func() {
		srv.Close()
	})
	wrapper := endpoint.DefaultWrapper(t.Logger())
	return &MockServer{
		wrapper: wrapper,
		srv:     srv,
		router:  router,
	}
}

func (m *MockServer) Client(opts ...httpcli.Option) *httpcli.Client {
	opts = append(opts, httpcli.WithGlobalRequestConfig(httpcli.GlobalRequestConfig{
		BaseUrl: m.BaseURL(),
	}))
	return httpcli.NewWithClient(m.srv.Client(), opts...)
}

func (m *MockServer) BaseURL() string {
	return m.srv.URL
}

func (m *MockServer) POST(path string, handler interface{}) *MockServer {
	return m.Mock(http.MethodPost, path, handler)
}

func (m *MockServer) GET(path string, handler interface{}) *MockServer {
	return m.Mock(http.MethodGet, path, handler)
}

func (m *MockServer) Mock(method string, path string, handler interface{}) *MockServer {
	m.router.Handler(method, path, m.wrapper.Endpoint(handler))
	return m
}
