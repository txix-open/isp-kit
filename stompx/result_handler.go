package stompx

import (
	"context"

	"github.com/go-stomp/stomp/v3"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/stompx/consumer"
)

type Result struct {
	ack     bool
	requeue bool
	err     error
}

type HandlerAdapter interface {
	Handle(ctx context.Context, msg *stomp.Message) Result
}

type AdapterFunc func(ctx context.Context, msg *stomp.Message) Result

func (a AdapterFunc) Handle(ctx context.Context, msg *stomp.Message) Result {
	return a(ctx, msg)
}

func Ack() Result {
	return Result{ack: true}
}

func Requeue(err error) Result {
	return Result{requeue: true, err: err}
}

type ResultHandler struct {
	logger  log.Logger
	adapter HandlerAdapter
}

func NewResultHandler(logger log.Logger, adapter HandlerAdapter) ResultHandler {
	return ResultHandler{
		logger:  logger,
		adapter: adapter,
	}
}

func (r ResultHandler) Handle(ctx context.Context, delivery *consumer.Delivery) {
	result := r.adapter.Handle(ctx, delivery.Source())

	switch {
	case result.ack:
		r.logger.Debug(ctx, "stomp client: message will be acknowledged")
		err := delivery.Ack()
		if err != nil {
			r.logger.Error(ctx, "stomp client: ack message error", log.Any("error", err))
		}
	case result.requeue:
		r.logger.Error(
			ctx,
			"stomp client: message will be requeued",
			log.Any("error", result.err),
		)
		err := delivery.Nack()
		if err != nil {
			r.logger.Error(ctx, "stomp client: nack message error", log.Any("error", err))
		}
	}
}
