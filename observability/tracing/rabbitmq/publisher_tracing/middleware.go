package publisher_tracing

import (
	"context"
	"fmt"

	"github.com/integration-system/grmq/publisher"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/observability/tracing/rabbitmq"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "isp-kit/observability/tracing/rabbitmq"
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
