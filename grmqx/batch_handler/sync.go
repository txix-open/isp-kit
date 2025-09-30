package batch_handler

import (
	"github.com/txix-open/isp-kit/log"
)

type SyncHandlerAdapter interface {
	Handle(items []Item) *Result
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(items []Item) *Result

func (a SyncHandlerAdapterFunc) Handle(items []Item) *Result {
	return a(items)
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

func (r Sync) Handle(items []Item) {
	if len(items) == 0 {
		return
	}

	result := r.handler.Handle(items)

	for _, ack := range result.ToAck {
		item := items[ack.Idx]
		err := item.Delivery.Ack()
		if err != nil {
			r.logger.Error(item.Context, "rmq client: ack message error", log.Any("error", err))
		}
	}

	for _, retry := range result.ToRetry {
		item := items[retry.Idx]
		err := item.Delivery.Retry()
		if err != nil {
			r.logger.Error(item.Context, "rmq client: retry message error", log.Any("error", err))
		}
	}

	for _, dlq := range result.ToDlq {
		item := items[dlq.Idx]
		err := item.Delivery.Nack(false)
		if err != nil {
			r.logger.Error(item.Context, "rmq client: nack message error", log.Any("error", err))
		}
	}
}
