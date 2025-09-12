package grpct

import (
	"net"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/test"
)

type MockServer struct {
	Wrapper endpoint.Wrapper
	srv     *grpc.Server
	router  *grpc.Mux
}

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

func (m *MockServer) Mock(endpoint string, handler any) *MockServer {
	m.router.Handle(endpoint, m.Wrapper.Endpoint(handler))
	return m
}

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
