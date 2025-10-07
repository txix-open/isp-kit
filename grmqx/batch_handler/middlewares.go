package batch_handler

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/panic_recovery"
)

type ConsumerMetricStorage interface {
	ObserveConsumeDuration(exchange string, routingKey string, t time.Duration)
	ObserveConsumeMsgSize(exchange string, routingKey string, size int)
	IncDlqCount(exchange string, routingKey string)
	IncSuccessCount(exchange string, routingKey string)
	IncRetryCount(exchange string, routingKey string)
}

func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []*BatchItem) {
			start := time.Now()
			next.Handle(items)

			for _, item := range items {
				exchange := item.Delivery.Source().Exchange
				routingKey := item.Delivery.Source().RoutingKey
				metricStorage.ObserveConsumeDuration(exchange, routingKey, time.Since(start))
				metricStorage.ObserveConsumeMsgSize(exchange, routingKey, len(item.Delivery.Source().Body))

				switch {
				case item.Result.Ack:
					metricStorage.IncSuccessCount(exchange, routingKey)
				case item.Result.Retry:
					metricStorage.IncRetryCount(exchange, routingKey)
				case item.Result.MoveToDlq:
					metricStorage.IncDlqCount(exchange, routingKey)
				}
			}
		})
	}
}

func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []*BatchItem) {
			next.Handle(items)

			for _, item := range items {
				exchange := item.Delivery.Source().Exchange
				routingKey := item.Delivery.Source().RoutingKey

				switch {
				case item.Result.Ack:
					logger.Debug(item.Context, "rmq client: batch message will be acknowledged",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
					)
				case item.Result.Retry:
					logger.Error(item.Context, "rmq client: batch message will be retried",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
						log.Any("error", item.Result.Err),
					)
				case item.Result.MoveToDlq:
					logger.Error(item.Context, "rmq client: batch message will be moved to DLQ or dropped",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
						log.Any("error", item.Result.Err),
					)
				}
			}
		})
	}
}

// nolint:nonamedreturns
func Recovery(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []*BatchItem) {
			defer panic_recovery.Recover(func(err error) {
				logger.Error(context.Background(), "rmq client: handle batch", log.Any("panic", err))
			})
			next.Handle(items)
		})
	}
}
