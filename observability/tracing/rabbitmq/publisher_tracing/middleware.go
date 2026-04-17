// Package publisher_tracing provides RabbitMQ publisher middleware for distributed tracing.
package publisher_tracing

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/observability/tracing/rabbitmq"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// tracerName identifies the tracer for RabbitMQ publisher tracing.
const tracerName = "isp-kit/observability/tracing/rabbitmq"

// Config holds the configuration for the RabbitMQ publisher tracing middleware.
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

// Middleware returns a RabbitMQ publisher middleware that creates spans for outgoing messages.
// It injects trace context into message headers, creates a producer span, and records
// delivery status. If the provider is a no-op, it returns a pass-through middleware.
func (c Config) Middleware() publisher.Middleware {
	if tracing.IsNoop(c.Provider) {
		return func(next publisher.RoundTripper) publisher.RoundTripper {
			return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
				return next.Publish(ctx, exchange, routingKey, msg)
			})
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			attributes := []attribute.KeyValue{
				tracing.RequestId.String(requestid.FromContext(ctx)),
				semconv.MessagingSystem("rabbitmq"),
				semconv.MessagingOperationPublish,
				semconv.MessagingDestinationName(exchange),
				semconv.MessagingRabbitmqDestinationRoutingKey(routingKey),
			}

			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindProducer),
				trace.WithAttributes(attributes...),
			}

			destination := rabbitmq.Destination(exchange, routingKey)
			spanName := fmt.Sprintf("%s %s", destination, semconv.MessagingOperationPublish.Value.AsString())

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			if msg.Headers == nil {
				msg.Headers = amqp091.Table{}
			}
			c.Propagator.Inject(ctx, rabbitmq.TableCarrier(msg.Headers))

			err := next.Publish(ctx, exchange, routingKey, msg)
			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		})
	}
}
