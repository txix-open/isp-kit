// Package client_tracing provides gRPC client middleware for distributed tracing.
package client_tracing

import (
	"context"
	"fmt"

	"github.com/txix-open/isp-kit/grpc/client/request"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/observability/tracing/grpc"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// tracerName identifies the tracer for gRPC client tracing.
const tracerName = "isp-kit/observability/tracing/grpc"

// Config holds the configuration for the gRPC client tracing middleware.
type Config struct {
	// Provider is the tracer provider used to create tracers.
	Provider tracing.TracerProvider
	// Propagator is the text map propagator for context propagation.
	Propagator tracing.Propagator
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

// Middleware returns a gRPC client middleware that creates spans for outgoing RPC calls.
// It injects trace context into the request metadata, creates a client span, and records
// errors. If the provider is a no-op, it returns a pass-through middleware.
func (c Config) Middleware() request.Middleware {
	if tracing.IsNoop(c.Provider) {
		return func(next request.RoundTripper) request.RoundTripper {
			return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
				return next(ctx, builder, message)
			}
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			attributes := []attribute.KeyValue{
				tracing.RequestId.String(requestid.FromContext(ctx)),
			}
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attributes...),
			}

			clientEndpoint := builder.Endpoint
			spanName := fmt.Sprintf("GRPC call %s", clientEndpoint)

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			c.Propagator.Inject(ctx, grpc.MetadataCarrier(builder.MD))

			resp, err := next(ctx, builder, message)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}

			return resp, err
		}
	}
}
