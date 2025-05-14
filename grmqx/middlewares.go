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

func PublisherLog(logger log.Logger, logBody bool) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			fields := []log.Field{
				log.String("exchange", exchange),
				log.String("routingKey", routingKey),
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

type Retrier interface {
	Do(ctx context.Context, f func() error) error
}

// PublisherRetry creates a middleware for retrying message publications.
// It is recommended to use this middleware after logging,
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

type PublisherMetricStorage interface {
	ObservePublishDuration(exchange string, routingKey string, t time.Duration)
	ObservePublishMsgSize(exchange string, routingKey string, size int)
	IncPublishError(exchange string, routingKey string)
}

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

func ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			fields := []log.Field{
				log.String("exchange", delivery.Source().Exchange),
				log.String("routingKey", delivery.Source().RoutingKey),
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
