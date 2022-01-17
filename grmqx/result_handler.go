package grmqx

import (
	"context"
	"time"

	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/isp-kit/log"
)

type Result struct {
	ack            bool
	requeue        bool
	requeueTimeout time.Duration
	moveToDlq      bool
	err            error
}

type HandlerAdapter interface {
	Handle(ctx context.Context, body []byte) Result
}

type AdapterFunc func(ctx context.Context, body []byte) Result

func (a AdapterFunc) Handle(ctx context.Context, body []byte) Result {
	return a(ctx, body)
}

func Ack() Result {
	return Result{ack: true}
}

func Requeue(after time.Duration, err error) Result {
	return Result{requeue: true, requeueTimeout: after, err: err}
}

// MoveToDlq
// if there is no DLQ, the message will be dropped
func MoveToDlq(err error) Result {
	return Result{moveToDlq: true, err: err}
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
	result := r.adapter.Handle(ctx, delivery.Source().Body)

	switch {
	case result.ack:
		r.logger.Debug(ctx, "rmq client: message will be acknowledged")
		err := delivery.Ack()
		if err != nil {
			r.logger.Error(ctx, "rmq client: ack message error", log.Any("error", err))
		}
	case result.requeue:
		r.logger.Error(
			ctx,
			"rmq client: message will be requeued",
			log.Any("error", result.err),
			log.String("requeueTimeout", result.requeueTimeout.String()),
		)
		select {
		case <-time.After(result.requeueTimeout):
		case <-ctx.Done():
		}
		err := delivery.Nack(true)
		if err != nil {
			r.logger.Error(ctx, "rmq client: nack message error", log.Any("error", err))
		}
	case result.moveToDlq:
		r.logger.Error(
			ctx,
			"rmq client: message will be moved to DLQ or dropped",
			log.Any("error", result.err),
		)
		err := delivery.Nack(false)
		if err != nil {
			r.logger.Error(ctx, "rmq client: nack message error", log.Any("error", err))
		}
	}
}
