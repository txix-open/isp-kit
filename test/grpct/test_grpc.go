package grpct

import (
	"net"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/client"
	endpoint2 "github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/test"
)

type MockServer struct {
	srv           *grpc.Server
	logger        log.Logger
	mockEndpoints map[string]any
}

func NewMock(t *test.Test) (*MockServer, *client.Client) {
	srv, cli := TestServer(t, grpc.NewMux())
	return &MockServer{
		srv:           srv,
		logger:        t.Logger(),
		mockEndpoints: make(map[string]any),
	}, cli
}

func (m *MockServer) Mock(endpoint string, handler any) *MockServer {
	m.mockEndpoints[endpoint] = handler
	wrapper := endpoint2.DefaultWrapper(m.logger)
	muxer := grpc.NewMux()
	for e, handler := range m.mockEndpoints {
		muxer.Handle(e, wrapper.Endpoint(handler))
	}
	m.srv.Upgrade(muxer)
	return m
}

func TestServer(t *test.Test, service isp.BackendServiceServer) (*grpc.Server, *client.Client) {
	assert := t.Assert()

	listener, err := net.Listen("tcp", "127.0.0.1:")
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
