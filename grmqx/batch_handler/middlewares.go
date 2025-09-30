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
	IncRequeueCount(exchange string, routingKey string)
	IncDlqCount(exchange string, routingKey string)
	IncSuccessCount(exchange string, routingKey string)
	IncRetryCount(exchange string, routingKey string)
}

func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []Item) *Result {
			start := time.Now()
			result := next.Handle(items)

			setMetric := func(idx int) (string, string) {
				exchange := items[idx].Delivery.Source().Exchange
				routingKey := items[idx].Delivery.Source().RoutingKey

				metricStorage.ObserveConsumeDuration(exchange, routingKey, time.Since(start))
				metricStorage.ObserveConsumeMsgSize(exchange, routingKey, len(items[idx].Delivery.Source().Body))
				return exchange, routingKey
			}

			for _, ack := range result.ToAck {
				exchange, routingKey := setMetric(ack.Idx)
				metricStorage.IncSuccessCount(exchange, routingKey)
			}

			for _, retry := range result.ToRetry {
				exchange, routingKey := setMetric(retry.Idx)
				metricStorage.IncRetryCount(exchange, routingKey)
			}

			for _, dlq := range result.ToDlq {
				exchange, routingKey := setMetric(dlq.Idx)
				metricStorage.IncDlqCount(exchange, routingKey)
			}

			return result
		})
	}
}

func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []Item) *Result {
			result := next.Handle(items)

			for _, ack := range result.ToAck {
				exchange := items[ack.Idx].Delivery.Source().Exchange
				routingKey := items[ack.Idx].Delivery.Source().RoutingKey

				logger.Debug(
					items[ack.Idx].Context,
					"rmq client: batch message will be acknowledged",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
				)
			}

			for _, retry := range result.ToRetry {
				exchange := items[retry.Idx].Delivery.Source().Exchange
				routingKey := items[retry.Idx].Delivery.Source().RoutingKey

				logger.Error(
					items[retry.Idx].Context,
					"rmq client: batch message will be retried",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.Any("error", retry.Err),
				)
			}

			for _, dlq := range result.ToDlq {
				exchange := items[dlq.Idx].Delivery.Source().Exchange
				routingKey := items[dlq.Idx].Delivery.Source().RoutingKey

				logger.Error(
					items[dlq.Idx].Context,
					"rmq client: batch message will be moved to DLQ or dropped",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.Any("error", dlq.Err),
				)
			}

			return result
		})
	}
}

// nolint:nonamedreturns
func Recovery(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(items []Item) *Result {
			defer panic_recovery.Recover(func(err error) {
				logger.Error(context.Background(), "rmq client: handle batch", log.Any("panic", err))
			})
			return next.Handle(items)
		})
	}
}
