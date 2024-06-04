package consumer_tracing

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/consumer"
	"gitlab.txix.ru/isp/isp-kit/grmqx/handler"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/rabbitmq"
	"gitlab.txix.ru/isp/isp-kit/requestid"
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
			/*opts = append(opts, trace.WithNewRoot())
			// Linking incoming span context if any for public endpoint.
			if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
				opts = append(opts, trace.WithLinks(trace.Link{SpanContext: s}))
			}*/

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
