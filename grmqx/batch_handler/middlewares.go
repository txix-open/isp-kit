package batch_handler

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/panic_recovery"
)

// ConsumerMetricStorage defines an interface for consumer metrics storage.
type ConsumerMetricStorage interface {
	ObserveConsumeDuration(exchange string, routingKey string, t time.Duration)
	ObserveConsumeMsgSize(exchange string, routingKey string, size int)
	IncDlqCount(exchange string, routingKey string)
	IncSuccessCount(exchange string, routingKey string)
	IncRetryCount(exchange string, routingKey string)
}

// Metrics creates a middleware that collects consumer metrics for batch processing,
// including message processing duration, message size, and result counts (success, retry, DLQ).
func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(batch BatchItems) {
			start := time.Now()
			next.Handle(batch)

			for _, item := range batch {
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

// Log creates a middleware that logs batch message processing results with appropriate log levels
// based on the outcome (Ack, Retry, or MoveToDlq).
func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(batch BatchItems) {
			next.Handle(batch)

			for _, item := range batch {
				exchange := item.Delivery.Source().Exchange
				routingKey := item.Delivery.Source().RoutingKey

				switch {
				case item.Result.Ack:
					logger.Debug(item.Context, "rmq client: batch message will be acknowledged",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
					)
				case item.Result.Retry && item.Result.Err != nil:
					logger.Error(item.Context, "rmq client: batch message will be retried",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
						log.Any("error", item.Result.Err),
					)
				case item.Result.Retry:
					logger.Debug(item.Context, "rmq client: batch message will be retried",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
					)
				case item.Result.MoveToDlq:
					logger.Error(item.Context, "rmq client: batch message will be moved to DLQ or dropped",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
						log.Any("error", item.Result.Err),
					)
				default:
					logger.Error(item.Context, "rmq client: batch message will be moved to DLQ or dropped",
						log.String("exchange", exchange),
						log.String("routingKey", routingKey),
						log.Any("error", errors.New("unexpected message result")),
					)
				}
			}
		})
	}
}

// Recovery creates a middleware that recovers from panics during batch message processing.
// On panic, the error is logged but message acknowledgment state remains unchanged.
func Recovery(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(batch BatchItems) {
			defer panic_recovery.Recover(func(err error) {
				logger.Error(context.Background(), "rmq client: handle batch", log.Any("panic", err))
			})
			next.Handle(batch)
		})
	}
}
