package handler

import (
	"context"
	"time"

	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/isp-kit/log"
)

type SyncHandlerAdapter interface {
	Handle(ctx context.Context, delivery *consumer.Delivery) Result
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(ctx context.Context, delivery *consumer.Delivery) Result

func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, delivery *consumer.Delivery) Result {
	return a(ctx, delivery)
}

type Sync struct {
	logger  log.Logger
	handler SyncHandlerAdapter
}

func NewSync(logger log.Logger, adapter SyncHandlerAdapter, middlewares ...Middleware) Sync {
	s := Sync{
		logger: logger,
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	s.handler = adapter
	return s
}

func (r Sync) Handle(ctx context.Context, delivery *consumer.Delivery) {
	result := r.handler.Handle(ctx, delivery)

	switch {
	case result.Ack:
		err := delivery.Ack()
		if err != nil {
			r.logger.Error(ctx, "rmq client: ack message error", log.Any("error", err))
		}
	case result.Requeue:
		select {
		case <-time.After(result.RequeueTimeout):
		case <-ctx.Done():
		}
		err := delivery.Nack(true)
		if err != nil {
			r.logger.Error(ctx, "rmq client: nack message error", log.Any("error", err))
		}
	case result.Retry:
		err := delivery.Retry()
		if err != nil {
			r.logger.Error(ctx, "rmq client: retry message error", log.Any("error", err))
		}
	case result.MoveToDlq:
		err := delivery.Nack(false)
		if err != nil {
			r.logger.Error(ctx, "rmq client: nack message error", log.Any("error", err))
		}
	}
}
