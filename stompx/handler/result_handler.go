// Package handler provides functionality for processing STOMP messages and handling results.
package handler

import (
	"context"
	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

// HandlerAdapter defines the interface for adapting message processing logic.
type HandlerAdapter interface {
	Handle(ctx context.Context, msg *stomp.Message) Result
}

// AdapterFunc is an adapter type that allows using functions as handler adapters.
type AdapterFunc func(ctx context.Context, msg *stomp.Message) Result

// Handle calls the underlying function.
func (a AdapterFunc) Handle(ctx context.Context, msg *stomp.Message) Result {
	return a(ctx, msg)
}

// ResultHandler wraps a handler adapter with middleware support and result processing.
type ResultHandler struct {
	logger  log.Logger
	adapter HandlerAdapter
}

// NewHandler creates a new ResultHandler with the provided logger, adapter, and optional middleware.
func NewHandler(logger log.Logger, adapter HandlerAdapter, middlewares ...Middleware) ResultHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	return ResultHandler{
		logger:  logger,
		adapter: adapter,
	}
}

// Handle processes a message delivery based on the adapter's result.
func (r ResultHandler) Handle(ctx context.Context, delivery *consumer.Delivery) {
	result := r.adapter.Handle(ctx, delivery.Source())

	switch {
	case result.Ack:
		err := delivery.Ack()
		if err != nil {
			r.logger.Error(ctx, "stomp client: ack message error", log.Any("error", err))
		}
	case result.Requeue:
		err := delivery.Nack()
		if err != nil {
			r.logger.Error(ctx, "stomp client: nack message error", log.Any("error", err))
		}
	}
}
