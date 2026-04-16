package request

import (
	"context"

	"github.com/txix-open/isp-kit/grpc/isp"
)

// RoundTripper defines the interface for executing gRPC requests.
// Receives a context, request builder, and message; returns a response message and optional error.
// Used as the base for middleware chains.
type RoundTripper func(ctx context.Context, builder *Builder, message *isp.Message) (*isp.Message, error)

// Middleware wraps a RoundTripper to add cross-cutting concerns like logging,
// metrics, authentication, or error handling.
// Middleware functions are typically chained in reverse order of execution.
type Middleware func(next RoundTripper) RoundTripper
