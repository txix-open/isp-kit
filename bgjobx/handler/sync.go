package handler

import (
	"context"

	"github.com/txix-open/bgjob"
)

type SyncHandlerAdapter interface {
	Handle(ctx context.Context, job bgjob.Job) Result
}

type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

type SyncHandlerAdapterFunc func(ctx context.Context, job bgjob.Job) Result

func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, job bgjob.Job) Result {
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
	result := r.handler.Handle(ctx, job)

	switch {
	case result.Complete:
		return bgjob.Complete()
	case result.Retry:
		return bgjob.Retry(result.RetryDelay, result.Err)
	case result.MoveToDlq:
		return bgjob.MoveToDlq(result.Err)
	case result.Reschedule:
		return bgjob.Reschedule(result.RescheduleDelay)
	case result.OverrideArg:
		return bgjob.RescheduleWithArg(result.RescheduleDelay, result.Arg)
	}

	return bgjob.Result{}
}
