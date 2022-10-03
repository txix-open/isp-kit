package grpc

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	DefaultMaxSizeByte = 64 * 1024 * 1024
)

type service struct {
	delegate atomic.Value
}

func (s *service) Request(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	delegate, ok := s.delegate.Load().(isp.BackendServiceServer)
	if !ok {
		return nil, errors.New("handler is not initialized")
	}
	return delegate.Request(ctx, message)
}

func (s *service) RequestStream(server isp.BackendService_RequestStreamServer) error {
	delegate, ok := s.delegate.Load().(isp.BackendServiceServer)
	if !ok {
		return errors.New("handler is not initialized")
	}
	return delegate.RequestStream(server)
}

type Server struct {
	server  *grpc.Server
	service *service
}

func DefaultServer(restOptions ...grpc.ServerOption) *Server {
	opts := append([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultMaxSizeByte),
		grpc.MaxSendMsgSize(DefaultMaxSizeByte),
	}, restOptions...)
	return NewServer(opts...)
}

func NewServer(opts ...grpc.ServerOption) *Server {
	s := &Server{
		server: grpc.NewServer(opts...),
		service: &service{
			delegate: atomic.Value{},
		},
	}
	isp.RegisterBackendServiceServer(s.server, s.service)
	return s
}

func (s *Server) Shutdown() {
	s.server.GracefulStop()
}

func (s *Server) Upgrade(service isp.BackendServiceServer) {
	s.service.delegate.Store(service)
}

func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.WithMessagef(err, "listen: %s", address)
	}
	return s.Serve(listener)
}

func (s *Server) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)
	if err != nil {
		return errors.WithMessage(err, "serve grpc")
	}
	return nil
}
