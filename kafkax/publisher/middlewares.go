package publisher

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
)

type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, partition int, offset int64, t time.Duration)
	ObservePublishMsgSize(topic string, partition int, offset int64, size int)
	IncPublishError(topic string, partition int, offset int64)
}

func PublisherMetrics(storage PublisherMetricStorage) Middleware {
	return func(next SyncPublisherAdapter) SyncPublisherAdapter {
		return SyncPublisherAdapterFunc(func(ctx context.Context, msg *kafka.Message) error {
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

func PublisherLog(logger log.Logger, w *kafka.Writer, connId string) Middleware {
	return func(next SyncPublisherAdapter) SyncPublisherAdapter {
		return SyncPublisherAdapterFunc(func(ctx context.Context, msg *kafka.Message) error {
			logger.Debug(
				ctx,
				"kafka client: publish message",
				log.String("addr", w.Addr.String()),
				log.String("topic", w.Topic),
				log.String("connId", connId),
				log.ByteString("messageKey", msg.Key),
				log.ByteString("messageValue", msg.Value),
			)

			return next.Publish(ctx, msg)
		})
	}
}
