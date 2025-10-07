package batch_handler

import (
	"github.com/txix-open/isp-kit/log"
)

type SyncHandlerAdapter interface {
	Handle(items []*BatchItem)
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(items []*BatchItem)

func (a SyncHandlerAdapterFunc) Handle(items []*BatchItem) {
	a(items)
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

func (r Sync) Handle(items []*BatchItem) {
	if len(items) == 0 {
		return
	}

	r.handler.Handle(items)

	for _, item := range items {
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
		}
	}
}
