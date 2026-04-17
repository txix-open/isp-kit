// Package handler provides synchronous message processing for RabbitMQ consumers.
package handler

import (
	"context"
	"time"

	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/log"
)

// SyncHandlerAdapter defines the interface for synchronous message handlers.
type SyncHandlerAdapter interface {
	Handle(ctx context.Context, delivery *consumer.Delivery) Result
}

// Middleware is a function that wraps a SyncHandlerAdapter.
type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

// SyncHandlerAdapterFunc is an adapter that allows using functions as SyncHandlerAdapters.
type SyncHandlerAdapterFunc func(ctx context.Context, delivery *consumer.Delivery) Result

// Handle handles a message delivery using the function.
func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, delivery *consumer.Delivery) Result {
	return a(ctx, delivery)
}

// Sync wraps a handler with middleware and manages message acknowledgment.
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

// Handle processes a message delivery and performs the appropriate action
// based on the Result (Ack, Requeue, Retry, or MoveToDlq).
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
