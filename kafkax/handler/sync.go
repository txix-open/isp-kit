package handler

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
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
	for {
		result := r.handler.Handle(ctx, delivery)
		switch {
		case result.Commit:
			err := delivery.Commit(ctx)
			if err != nil {
				r.logger.Error(
					ctx, "kafka consumer: unexpected error during committing message",
					log.Any("error", err),
				)
			}
			return
		case result.Retry:
			select {
			case <-ctx.Done():
				delivery.Done()
				return
			case <-time.After(result.RetryAfter):
				continue
			}
		}
	}
}
