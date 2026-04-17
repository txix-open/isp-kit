package handler

import (
	"context"
	"time"

	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/panic_recovery"
)

// ConsumerMetricStorage defines an interface for consumer metrics storage.
type ConsumerMetricStorage interface {
	ObserveConsumeDuration(exchange string, routingKey string, t time.Duration)
	ObserveConsumeMsgSize(exchange string, routingKey string, size int)
	IncRequeueCount(exchange string, routingKey string)
	IncDlqCount(exchange string, routingKey string)
	IncSuccessCount(exchange string, routingKey string)
	IncRetryCount(exchange string, routingKey string)
}

// Metrics creates a middleware that collects consumer metrics including
// message processing duration, message size, and result counts (success, requeue, retry, DLQ).
func Metrics(metricStorage ConsumerMetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) Result {
			exchange := delivery.Source().Exchange
			routingKey := delivery.Source().RoutingKey
			start := time.Now()
			result := next.Handle(ctx, delivery)
			metricStorage.ObserveConsumeDuration(exchange, routingKey, time.Since(start))
			metricStorage.ObserveConsumeMsgSize(exchange, routingKey, len(delivery.Source().Body))

			switch {
			case result.Ack:
				metricStorage.IncSuccessCount(exchange, routingKey)
			case result.Requeue:
				metricStorage.IncRequeueCount(exchange, routingKey)
			case result.Retry:
				metricStorage.IncRetryCount(exchange, routingKey)
			case result.MoveToDlq:
				metricStorage.IncDlqCount(exchange, routingKey)
			}

			return result
		})
	}
}

// Log creates a middleware that logs message processing results with appropriate log levels
// based on the outcome (Ack, Requeue, Retry, or MoveToDlq).
func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) Result {
			exchange := delivery.Source().Exchange
			routingKey := delivery.Source().RoutingKey

			result := next.Handle(ctx, delivery)

			switch {
			case result.Ack:
				logger.Debug(
					ctx,
					"rmq client: message will be acknowledged",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
				)
			case result.Requeue && result.Err != nil:
				logger.Error(
					ctx,
					"rmq client: message will be requeued",
					log.Any("error", result.Err),
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.String("requeueTimeout", result.RequeueTimeout.String()),
				)
			case result.Requeue:
				logger.Debug(
					ctx,
					"rmq client: message will be requeued",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.String("requeueTimeout", result.RequeueTimeout.String()),
				)
			case result.Retry && result.Err != nil:
				logger.Error(
					ctx,
					"rmq client: message will be retried",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.Any("error", result.Err),
				)
			case result.Retry:
				logger.Debug(
					ctx,
					"rmq client: message will be retried",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
				)
			case result.MoveToDlq:
				logger.Error(
					ctx,
					"rmq client: message will be moved to DLQ or dropped",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.Any("error", result.Err),
				)
			}

			return result
		})
	}
}

// Recovery creates a middleware that recovers from panics during message processing.
// On panic, the message is moved to the DLQ and the error is recorded.
func Recovery() Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) (res Result) {
			defer panic_recovery.Recover(func(err error) {
				res.Err = err
				res.MoveToDlq = true
			})
			return next.Handle(ctx, delivery)
		})
	}
}
