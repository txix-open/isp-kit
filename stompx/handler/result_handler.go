package handler

import (
	"context"
	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

type HandlerAdapter interface {
	Handle(ctx context.Context, msg *stomp.Message) Result
}

type AdapterFunc func(ctx context.Context, msg *stomp.Message) Result

func (a AdapterFunc) Handle(ctx context.Context, msg *stomp.Message) Result {
	return a(ctx, msg)
}

type ResultHandler struct {
	logger  log.Logger
	adapter HandlerAdapter
}

func NewHandler(logger log.Logger, adapter HandlerAdapter, middlewares ...Middleware) ResultHandler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	return ResultHandler{
		logger:  logger,
		adapter: adapter,
	}
}

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
