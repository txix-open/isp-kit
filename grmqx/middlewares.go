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

const (
	RequestIdHeader = "x-request-id"
)

func PublisherLog(logger log.Logger) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			logger.Debug(
				ctx,
				"rmq client: publish message",
				log.String("exchange", exchange),
				log.String("routingKey", routingKey),
				log.ByteString("body", msg.Body),
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
			msg.Headers[RequestIdHeader] = requestId
			return next.Publish(ctx, exchange, routingKey, msg)
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

func ConsumerLog(logger log.Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			logger.Debug(
				ctx,
				"rmq client: consume message",
				log.String("exchange", delivery.Source().Exchange),
				log.String("routingKey", delivery.Source().RoutingKey),
				log.ByteString("body", delivery.Source().Body),
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
				value, ok := headers[RequestIdHeader].(string)
				if ok {
					requestId = value
				}
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))
			next.Handle(ctx, delivery)
		})
	}
}
