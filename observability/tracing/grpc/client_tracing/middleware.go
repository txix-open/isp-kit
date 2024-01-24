package client_tracing

import (
	"context"
	"fmt"

	"github.com/integration-system/isp-kit/grpc/client/request"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/observability/tracing/grpc"
	"github.com/integration-system/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
