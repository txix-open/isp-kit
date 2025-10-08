package batch_handler

import (
	"github.com/txix-open/isp-kit/log"
)

type SyncHandlerAdapter interface {
	Handle(batch BatchItems)
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(batch BatchItems)

func (a SyncHandlerAdapterFunc) Handle(batch BatchItems) {
	a(batch)
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

func (r Sync) Handle(batch BatchItems) {
	if len(batch) == 0 {
		return
	}

	r.handler.Handle(batch)

	for _, item := range batch {
		switch {
		case item.Result.Ack:
			err := item.Delivery.Ack()
			if err != nil {
				r.logger.Error(item.Context, "rmq client: ack message error", log.Any("error", err))
			}
		case item.Result.Retry:
			err := item.Delivery.Retry()
			if err != nil {
				r.logger.Error(item.Context, "rmq client: retry message error", log.Any("error", err))
			}
		case item.Result.MoveToDlq:
			err := item.Delivery.Nack(false)
			if err != nil {
				r.logger.Error(item.Context, "rmq client: nack message error", log.Any("error", err))
			}
		default:
			err := item.Delivery.Nack(false)
			if err != nil {
				r.logger.Error(item.Context, "rmq client: nack message error", log.Any("error", err))
			}
		}
	}
}
