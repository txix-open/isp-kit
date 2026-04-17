package grmqx

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

// PublisherLog creates a publisher middleware that logs published messages.
// When logBody is true, the message body is included in the log output.
func PublisherLog(logger log.Logger, logBody bool) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			fields := []log.Field{
				log.String("exchange", exchange),
				log.String("routingKey", routingKey),
				log.Int("bodySize", len(msg.Body)),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", msg.Body))
			}
			logger.Debug(
				ctx,
				"rmq client: publish message",
				fields...,
			)
			return next.Publish(ctx, exchange, routingKey, msg)
		})
	}
}

// PublisherRequestId creates a publisher middleware that generates and injects request IDs
// into message headers. If a request ID already exists in the context, it is used; otherwise,
// a new one is generated.
func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}
			if msg.Headers == nil {
				msg.Headers = amqp091.Table{}
			}
			msg.Headers[requestid.Header] = requestId
			return next.Publish(ctx, exchange, routingKey, msg)
		})
	}
}

// Retrier defines an interface for retry logic.
type Retrier interface {
	Do(ctx context.Context, f func() error) error
}

// PublisherRetry creates a middleware for retrying message publications.
// It is recommended to use this middleware after logging middleware
// to avoid duplicate logging of publication attempts.
func PublisherRetry(retrier Retrier) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			return retrier.Do(ctx, func() error {
				return next.Publish(ctx, exchange, routingKey, msg)
			})
		})
	}
}

// PublisherMetricStorage defines an interface for publisher metrics storage.
type PublisherMetricStorage interface {
	ObservePublishDuration(exchange string, routingKey string, t time.Duration)
	ObservePublishMsgSize(exchange string, routingKey string, size int)
	IncPublishError(exchange string, routingKey string)
}

// PublisherMetrics creates a middleware that collects publisher metrics including
// message size, publication duration, and error counts.
func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			storage.ObservePublishMsgSize(exchange, routingKey, len(msg.Body))
			start := time.Now()
			err := next.Publish(ctx, exchange, routingKey, msg)
			if err != nil {
				storage.IncPublishError(exchange, routingKey)
			}
			storage.ObservePublishDuration(exchange, routingKey, time.Since(start))
			return err
		})
	}
}

// ConsumerLog creates a consumer middleware that logs consumed messages.
// When logBody is true, the message body is included in the log output.
func ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			fields := []log.Field{
				log.String("exchange", delivery.Source().Exchange),
				log.String("routingKey", delivery.Source().RoutingKey),
				log.Int("bodySize", len(delivery.Source().Body)),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", delivery.Source().Body))
			}
			logger.Debug(
				ctx,
				"rmq client: consume message",
				fields...,
			)
			next.Handle(ctx, delivery)
		})
	}
}

// ConsumerRequestId creates a consumer middleware that extracts request IDs from message headers
// and propagates them in the context. If no request ID is found, a new one is generated.
// It also adds the request ID to the context logger.
func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			requestId := ""
			headers := delivery.Source().Headers
			if headers != nil {
				value, ok := headers[requestid.Header].(string)
				if ok {
					requestId = value
				}
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))
			next.Handle(ctx, delivery)
		})
	}
}
