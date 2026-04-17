// Package httpt provides test helpers for HTTP server and client operations.
// It creates mock HTTP servers on random local ports for testing.
package httpt

import (
	"net/http"
	"net/http/httptest"

	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/router"
	"github.com/txix-open/isp-kit/test"
)

// MockServer provides a mock HTTP server for testing.
// It uses httptest.Server and allows registering handlers for specific
// HTTP methods and paths.
type MockServer struct {
	Wrapper endpoint.Wrapper
	srv     *httptest.Server
	router  *router.Router
}

// NewMock creates a new mock HTTP server.
// The server listens on a random local port and is automatically
// shut down when the test completes.
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

// Client creates a new HTTP client configured to send requests to the
// mock server's base URL. Optional client configuration can be provided.
func (m *MockServer) Client(opts ...httpcli.Option) *httpcli.Client {
	cli := httpcli.New(opts...)
	cli.GlobalRequestConfig().BaseUrl = m.BaseURL()
	return cli
}

// BaseURL returns the base URL of the mock server.
func (m *MockServer) BaseURL() string {
	return m.srv.URL
}

// POST registers a handler for POST requests at the specified path.
// Returns the MockServer for method chaining.
func (m *MockServer) POST(path string, handler any) *MockServer {
	return m.Mock(http.MethodPost, path, handler)
}

// GET registers a handler for GET requests at the specified path.
// Returns the MockServer for method chaining.
func (m *MockServer) GET(path string, handler any) *MockServer {
	return m.Mock(http.MethodGet, path, handler)
}

// Mock registers a handler for the specified HTTP method and path.
// Returns the MockServer for method chaining.
func (m *MockServer) Mock(method string, path string, handler any) *MockServer {
	m.router.Handler(method, path, m.Wrapper.Endpoint(handler))
	return m
}
