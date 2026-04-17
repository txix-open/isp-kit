// Package grpc provides a gRPC server and client implementation for the ISP kit framework.
// It enables building scalable microservices with support for hot-swappable handlers,
// middleware integration, and structured authentication via metadata.
//
// The package supports both unary and streaming RPCs, with built-in features for
// request/response logging, metrics collection, and distributed tracing.
package grpc

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"google.golang.org/grpc"
)

const (
	// DefaultMaxSizeByte defines the default maximum message size in bytes (64MB).
	DefaultMaxSizeByte = 64 * 1024 * 1024
)

// service implements the BackendServiceServer interface with hot-swappable handler support.
// It uses atomic.Value for thread-safe handler replacement without interrupting in-flight requests.
type service struct {
	isp.UnimplementedBackendServiceServer

	delegate atomic.Value
}

// Request handles unary gRPC requests by delegating to the registered BackendServiceServer.
// Returns an error if the handler is not initialized.
func (s *service) Request(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	delegate, ok := s.delegate.Load().(isp.BackendServiceServer)
	if !ok {
		return nil, errors.New("handler is not initialized")
	}
	return delegate.Request(ctx, message)
}

// RequestStream handles streaming gRPC requests by delegating to the registered BackendServiceServer.
// Returns an error if the handler is not initialized.
func (s *service) RequestStream(server isp.BackendService_RequestStreamServer) error {
	delegate, ok := s.delegate.Load().(isp.BackendServiceServer)
	if !ok {
		return errors.New("handler is not initialized")
	}
	return delegate.RequestStream(server)
}

// Server wraps a gRPC server with hot-swappable handler support.
// It provides a clean interface for starting, stopping, and upgrading gRPC services
// without requiring server restart.
type Server struct {
	server  *grpc.Server
	service *service
}

// DefaultServer creates a new Server with default configuration including 64MB message size limits.
// Accepts additional grpc.ServerOption for further customization.
// Thread-safe for concurrent use.
func DefaultServer(restOptions ...grpc.ServerOption) *Server {
	opts := append([]grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultMaxSizeByte),
		grpc.MaxSendMsgSize(DefaultMaxSizeByte),
	}, restOptions...)
	return NewServer(opts...)
}

// NewServer creates a new Server with custom gRPC options.
// Accepts variadic grpc.ServerOption for full configuration control.
// Thread-safe for concurrent use.
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

// Shutdown gracefully stops the server, waiting for in-flight requests to complete.
// Thread-safe for concurrent use.
func (s *Server) Shutdown() {
	s.server.GracefulStop()
}

// Upgrade atomically replaces the backend service handler without interrupting in-flight requests.
// This enables hot-reload of business logic without server restart.
// Thread-safe for concurrent use.
func (s *Server) Upgrade(service isp.BackendServiceServer) {
	s.service.delegate.Store(service)
}

// ListenAndServe binds the server to the specified TCP address and starts serving.
// Returns an error if binding fails or serving encounters a fatal error.
func (s *Server) ListenAndServe(address string) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", address)
	if err != nil {
		return errors.WithMessagef(err, "listen: %s", address)
	}
	return s.Serve(listener)
}

// Serve starts serving gRPC requests on the provided listener.
// Blocks until the server is stopped or an error occurs.
// Returns an error if serving encounters a fatal error.
func (s *Server) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)
	if err != nil {
		return errors.WithMessage(err, "serve grpc")
	}
	return nil
}
