package handler

import (
	"context"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/log"
)

const (
	panicStackLength = 4 << 10
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
			case result.Requeue:
				logger.Error(
					ctx,
					"rmq client: message will be requeued",
					log.Any("error", result.Err),
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.String("requeueTimeout", result.RequeueTimeout.String()),
				)
			case result.Retry:
				logger.Error(
					ctx,
					"rmq client: message will be retried",
					log.String("exchange", exchange),
					log.String("routingKey", routingKey),
					log.Any("error", result.Err),
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
				res.Err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
				res.MoveToDlq = true
			}()
			return next.Handle(ctx, delivery)
		})
	}
}
