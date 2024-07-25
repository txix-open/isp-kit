package handler

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
)

type ConsumerMetricStorage interface {
	ObserveConsumeDuration(topic string, partition int, offset int64, t time.Duration)
	ObserveConsumeMsgSize(topic string, partition int, offset int64, size int)
	IncCommitCount(topic string, partition int, offset int64)
	IncRetryCount(topic string, partition int, offset int64)
}

func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) Result {
			topic := msg.Topic
			partition := msg.Partition
			offset := msg.Offset
			start := time.Now()

			result := next.Handle(ctx, msg)

			metricStorage.ObserveConsumeDuration(topic, partition, offset, time.Since(start))
			metricStorage.ObserveConsumeMsgSize(topic, partition, offset, len(msg.Value))

			switch {
			case result.Commit:
				metricStorage.IncCommitCount(topic, partition, offset)
			case result.Retry:
				metricStorage.IncRetryCount(topic, partition, offset)
			}

			return result
		})
	}
}

func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) Result {
			topic := msg.Topic
			partition := msg.Partition
			offset := msg.Offset

			result := next.Handle(ctx, msg)

			switch {
			case result.Commit:
				logger.Debug(
					ctx,
					"kafka client: message will be commited",
					log.String("topic", topic),
					log.Int("partition", partition),
					log.Int64("offset", offset),
				)
			case result.Retry:
				logger.Error(
					ctx,
					"kafka client: message will be retried",
					log.String("topic", topic),
					log.Int("partition", partition),
					log.Int64("offset", offset),
					log.Any("error", result.RetryError),
					log.String("retryAfter", result.RetryAfter.String()),
				)
			}

			return result
		})
	}
}

func ConsumerLog(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) Result {
			logger.Debug(
				ctx,
				"kafka consumer: consume message",
				log.String("topic", msg.Topic),
				log.Int("partition", msg.Partition),
				log.Int64("offset", msg.Offset),
				log.ByteString("messageKey", msg.Key),
				log.ByteString("messageValue", msg.Value),
			)
			return next.Handle(ctx, msg)
		})
	}
}
