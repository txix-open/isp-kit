package handler

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
)

type SyncHandlerAdapter interface {
	Handle(ctx context.Context, msg *kafka.Message) Result
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(ctx context.Context, msg *kafka.Message) Result

func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, msg *kafka.Message) Result {
	return a(ctx, msg)
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

func (r Sync) Handle(ctx context.Context, msg *kafka.Message) Result {
	return r.handler.Handle(ctx, msg)
}
