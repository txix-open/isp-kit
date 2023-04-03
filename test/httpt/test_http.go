package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/integration-system/isp-kit/http/endpoint"
	"github.com/integration-system/isp-kit/test"
	"github.com/julienschmidt/httprouter"
)

type MockServer struct {
	wrapper endpoint.Wrapper
	srv     *httptest.Server
	router  *httprouter.Router
}

func NewMock(t *test.Test) *MockServer {
	router := httprouter.New()
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

func (m *MockServer) Client() *http.Client {
	return m.srv.Client()
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
