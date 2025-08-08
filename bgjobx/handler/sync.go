package handler

import (
	"context"

	"github.com/txix-open/bgjob"
)

type SyncHandlerAdapter interface {
	Handle(ctx context.Context, job bgjob.Job) bgjob.Result
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(ctx context.Context, job bgjob.Job) bgjob.Result

func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, job bgjob.Job) bgjob.Result {
	return a(ctx, job)
}

type Sync struct {
	handler SyncHandlerAdapter
}

func NewSync(adapter SyncHandlerAdapter, middlewares ...Middleware) Sync {
	s := Sync{}
	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	s.handler = adapter
	return s
}

func (r Sync) Handle(ctx context.Context, job bgjob.Job) bgjob.Result {
	return r.handler.Handle(ctx, job)
}
