// Package consumer_tracing provides RabbitMQ consumer middleware for distributed tracing.
package consumer_tracing

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx/handler"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/observability/tracing/rabbitmq"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// tracerName identifies the tracer for RabbitMQ consumer tracing.
const tracerName = "isp-kit/observability/tracing/rabbitmq"

// Config holds the configuration for the RabbitMQ consumer tracing middleware.
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

// Middleware returns a RabbitMQ consumer middleware that creates spans for incoming messages.
// It extracts trace context from message headers, creates a consumer span, and records
// the processing result (ack, requeue, retry, or dlq). If the provider is a no-op,
// it returns a pass-through middleware.
func (c Config) Middleware() handler.Middleware {
	if tracing.IsNoop(c.Provider) {
		return func(next handler.SyncHandlerAdapter) handler.SyncHandlerAdapter {
			return handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
				return next.Handle(ctx, delivery)
			})
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next handler.SyncHandlerAdapter) handler.SyncHandlerAdapter {
		return handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
			source := delivery.Source()
			if source.Headers == nil {
				source.Headers = amqp091.Table{}
			}
			ctx = c.Propagator.Extract(ctx, rabbitmq.TableCarrier(source.Headers))

			attributes := []attribute.KeyValue{
				tracing.RequestId.String(requestid.FromContext(ctx)),
				semconv.MessagingSystem("rabbitmq"),
				semconv.MessagingOperationKey.String("deliver"),
				semconv.MessagingDestinationName(source.Exchange),
				semconv.MessagingRabbitmqDestinationRoutingKey(source.RoutingKey),
			}
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(attributes...),
			}

			destination := rabbitmq.Destination(source.Exchange, source.RoutingKey)
			spanName := fmt.Sprintf("%s deliver", destination)

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			result := next.Handle(ctx, delivery)
			switch {
			case result.Ack:
				span.SetStatus(codes.Ok, "")
			case result.Requeue:
				span.RecordError(result.Err)
				span.SetStatus(codes.Error, errors.WithMessage(result.Err, "message will be requeued").Error())
			case result.Retry:
				span.RecordError(result.Err)
				span.SetStatus(codes.Error, errors.WithMessage(result.Err, "message will be retried").Error())
			case result.MoveToDlq:
				span.RecordError(result.Err)
				span.SetStatus(codes.Error, errors.WithMessage(result.Err, "message will be dropped").Error())
			}

			return result
		})
	}
}
