package grpc

import (
	"context"

	"github.com/txix-open/isp-kit/grpc/isp"
)

// HandlerFunc defines the signature for gRPC request handlers.
// Receives a context and message, returns a response message and optional error.
type HandlerFunc func(ctx context.Context, message *isp.Message) (*isp.Message, error)

// Middleware wraps a HandlerFunc to add cross-cutting concerns like logging,
// metrics, authentication, or error handling.
// Middleware functions are typically chained in reverse order of execution.
type Middleware func(next HandlerFunc) HandlerFunc
