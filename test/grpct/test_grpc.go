// Package grpct provides test helpers for gRPC server and client operations.
// It creates mock gRPC servers on random local ports for testing.
package grpct

import (
	"net"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/test"
)

// MockServer provides a mock gRPC server for testing.
// It allows registering handlers for specific endpoints.
type MockServer struct {
	Wrapper endpoint.Wrapper
	srv     *grpc.Server
	router  *grpc.Mux
}

// NewMock creates a new mock gRPC server and client pair.
// The server listens on a random local port and is automatically
// shut down when the test completes.
func NewMock(t *test.Test) (*MockServer, *client.Client) {
	srv, cli := TestServer(t, grpc.NewMux())
	router := grpc.NewMux()
	srv.Upgrade(router)
	return &MockServer{
		Wrapper: endpoint.DefaultWrapper(t.Logger()),
		srv:     srv,
		router:  router,
	}, cli
}

// Mock registers a handler for the specified endpoint.
// Returns the MockServer for method chaining.
func (m *MockServer) Mock(endpoint string, handler any) *MockServer {
	m.router.Handle(endpoint, m.Wrapper.Endpoint(handler))
	return m
}

// TestServer creates and starts a gRPC server with the provided service,
// returning both the server and a configured client. The server listens
// on a random local port and is automatically shut down when the test completes.
func TestServer(t *test.Test, service isp.BackendServiceServer) (*grpc.Server, *client.Client) {
	assert := t.Assert()

	var lc net.ListenConfig
	listener, err := lc.Listen(t.T().Context(), "tcp", "127.0.0.1:")
	assert.NoError(err)
	srv := grpc.NewServer()
	cli, err := client.Default()
	assert.NoError(err)
	t.T().Cleanup(func() {
		err := cli.Close()
		assert.NoError(err)
		srv.Shutdown()
	})
	srv.Upgrade(service)
	go func() {
		err := srv.Serve(listener)
		assert.NoError(err)
	}()

	cli.Upgrade([]string{listener.Addr().String()})
	return srv, cli
}
