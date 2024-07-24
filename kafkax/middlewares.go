package kafkax

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/log"
)

func PublisherLog(logger log.Logger) Middleware {
	return func(next PublisherMiddleware) PublisherMiddleware {
		return PublisherMiddlewareFunc(func(ctx context.Context, msg kafka.Message) error {
			logger.Debug(
				ctx,
				"kafka client: publish message",
				log.String("topic", msg.Topic),
				log.Int("partition", msg.Partition),
				log.Int64("offset", msg.Offset),
				log.ByteString("msg", msg.Value),
			)
			return next.Publish(ctx, msg)
		})
	}
}

type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, partition int, offset int64, t time.Duration)
	ObservePublishMsgSize(topic string, partition int, offset int64, size int)
	IncPublishError(topic string, partition int, offset int64)
}

func PublisherMetrics(storage PublisherMetricStorage) Middleware {
	return func(next PublisherMiddleware) PublisherMiddleware {
		return PublisherMiddlewareFunc(func(ctx context.Context, msg kafka.Message) error {
			storage.ObservePublishMsgSize(msg.Topic, msg.Partition, msg.Offset, len(msg.Value))
			start := time.Now()

			err := next.Publish(ctx, msg)
			if err != nil {
				storage.IncPublishError(msg.Topic, msg.Partition, msg.Offset)
			}

			storage.ObservePublishDuration(msg.Topic, msg.Partition, msg.Offset, time.Since(start))
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
