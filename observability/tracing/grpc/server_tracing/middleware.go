package server_tracing

import (
	"context"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/log/logutil"
	"github.com/integration-system/isp-kit/observability/tracing"
	grpc2 "github.com/integration-system/isp-kit/observability/tracing/grpc"
	"github.com/integration-system/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

const (
	tracerName = "isp-kit/observability/tracing/grpc"
)

type Config struct {
	Provider   tracing.TracerProvider
	Propagator tracing.Propagator
}

func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

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
