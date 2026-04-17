// Package server_tracing provides gRPC server middleware for distributed tracing.
package server_tracing

import (
	"context"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/logutil"
	"github.com/txix-open/isp-kit/observability/tracing"
	grpc2 "github.com/txix-open/isp-kit/observability/tracing/grpc"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// tracerName identifies the tracer for gRPC server tracing.
const tracerName = "isp-kit/observability/tracing/grpc"

// Config holds the configuration for the gRPC server tracing middleware.
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

// Middleware returns a gRPC server middleware that creates spans for incoming RPC calls.
// It extracts trace context from request metadata, creates a server span, and records
// errors at the appropriate log level. If the provider is a no-op, it returns a pass-through middleware.
func (c Config) Middleware() grpc.Middleware {
	if tracing.IsNoop(c.Provider) {
		return func(next grpc.HandlerFunc) grpc.HandlerFunc {
			return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
				return next(ctx, message)
			}
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			if md == nil {
				md = metadata.MD{}
			}
			spanName, _ := grpc.StringFromMd(grpc.ProxyMethodNameHeader, md)
			if spanName == "" {
				return next(ctx, message)
			}

			ctx = c.Propagator.Extract(ctx, grpc2.MetadataCarrier(md))

			attributes := []attribute.KeyValue{
				tracing.RequestId.String(requestid.FromContext(ctx)),
			}
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attributes...),
			}

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			resp, err := next(ctx, message)

			if err != nil {
				logLevel := logutil.LogLevelForError(err)
				if logLevel == log.ErrorLevel {
					span.RecordError(err)
				}
			}

			return resp, err
		}
	}
}
