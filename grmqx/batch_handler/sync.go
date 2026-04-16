package batch_handler

import (
	"github.com/txix-open/isp-kit/log"
)

// SyncHandlerAdapter defines the interface for synchronous batch message handlers.
type SyncHandlerAdapter interface {
	Handle(batch BatchItems)
}

// Middleware is a function that wraps a SyncHandlerAdapter.
type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

// SyncHandlerAdapterFunc is an adapter that allows using functions as SyncHandlerAdapters.
type SyncHandlerAdapterFunc func(batch BatchItems)

// Handle handles a batch of messages using the function.
func (a SyncHandlerAdapterFunc) Handle(batch BatchItems) {
	a(batch)
}

// Sync wraps a handler with middleware and manages batch message acknowledgment.
type Sync struct {
	logger  log.Logger
	handler SyncHandlerAdapter
}

// NewSync creates a new Sync handler with the specified logger, adapter, and middleware.
// Middleware functions are applied in reverse order (last to first).
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

// Handle processes a batch of messages and performs the appropriate action
// for each message based on its Result (Ack, Retry, or MoveToDlq).
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
