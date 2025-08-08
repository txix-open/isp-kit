package handler

import (
	"context"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
)

const (
	panicStackLength          = 4 << 10
	recoverMessageRetryPeriod = 1 * time.Hour
)

type ConsumerMetricStorage interface {
	ObserveConsumeDuration(consumerGroup, topic string, t time.Duration)
	ObserveConsumeMsgSize(consumerGroup, topic string, size int)
	IncCommitCount(consumerGroup, topic string)
	IncRetryCount(consumerGroup, topic string)
}

func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) Result {
			topic := delivery.Source().Topic
			consumerGroup := delivery.ConsumerGroupId()
			start := time.Now()

			result := next.Handle(ctx, delivery)

			metricStorage.ObserveConsumeDuration(consumerGroup, topic, time.Since(start))
			metricStorage.ObserveConsumeMsgSize(consumerGroup, topic, len(delivery.Source().Value))

			switch {
			case result.Commit:
				metricStorage.IncCommitCount(consumerGroup, topic)
			case result.Retry:
				metricStorage.IncRetryCount(consumerGroup, topic)
			}

			return result
		})
	}
}

func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) Result {
			topic := delivery.Source().Topic
			partition := delivery.Source().Partition
			offset := delivery.Source().Offset

			result := next.Handle(ctx, delivery)

			switch {
			case result.Commit:
				logger.Debug(
					ctx,
					"kafka client: message will be committed",
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
					log.String("error", result.RetryError.Error()),
					log.String("retryAfter", result.RetryAfter.String()),
				)
			}

			return result
		})
	}
}

// nolint:nonamedreturns
func Recovery() Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) (res Result) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				var err error
				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = errors.Errorf("%v", recovered)
				}
				stack := make([]byte, panicStackLength)
				length := runtime.Stack(stack, false)
				res.RetryError = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				res.Retry = true
				res.RetryAfter = recoverMessageRetryPeriod
			}()
			return next.Handle(ctx, delivery)
		})
	}
}
