package grpc

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// ProxyMethodNameHeader is the metadata key used to specify the target endpoint.
	ProxyMethodNameHeader = "proxy_method_name"
)

// Mux is a request router that dispatches gRPC requests to registered handlers based on endpoint name.
// It extracts the endpoint from request metadata and invokes the corresponding HandlerFunc.
// Mux implements the BackendServiceServer interface but does not support streaming RPCs.
type Mux struct {
	isp.UnimplementedBackendServiceServer

	unaryHandlers map[string]HandlerFunc
}

// NewMux creates a new Mux with an empty handler registry.
func NewMux() *Mux {
	return &Mux{
		unaryHandlers: make(map[string]HandlerFunc),
	}
}

// Handle registers a HandlerFunc for the specified endpoint.
// Panics if a handler is already registered for the endpoint.
// Returns the Mux for method chaining.
func (m *Mux) Handle(endpoint string, handler HandlerFunc) *Mux {
	_, ok := m.unaryHandlers[endpoint]
	if ok {
		panic(errors.Errorf("handler for endpoint %v is already provided", endpoint))
	}
	m.unaryHandlers[endpoint] = handler
	return m
}

// Request handles unary gRPC requests by routing to the registered handler for the endpoint.
// The endpoint is extracted from the ProxyMethodNameHeader in request metadata.
// Returns an error if metadata is missing, endpoint is not found, or handler fails.
func (m *Mux) Request(ctx context.Context, message *isp.Message) (*isp.Message, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("metadata is expected in context")
	}
	endpoint, err := StringFromMd(ProxyMethodNameHeader, md)
	if err != nil {
		return nil, err
	}
	handler, ok := m.unaryHandlers[endpoint]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "handler not found for endpoint %s", endpoint)
	}
	ctx = log.ToContext(ctx, log.String("endpoint", endpoint))
	return handler(ctx, message)
}

// RequestStream returns an unimplemented error for streaming RPCs.
// Mux does not support streaming; use a custom implementation for streaming support.
func (m *Mux) RequestStream(_ isp.BackendService_RequestStreamServer) error {
	return status.Errorf(codes.Unimplemented, "service does not support streaming RPC")
}
