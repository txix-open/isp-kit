package handler

import (
	"context"

	"github.com/txix-open/bgjob"
)

// SyncHandlerAdapter defines the interface for synchronous job handlers.
// Implementations must process a job and return a Result indicating the
// desired action (complete, retry, reschedule, or move to DLQ).
type SyncHandlerAdapter interface {
	Handle(ctx context.Context, job bgjob.Job) Result
}

// Middleware is a function that wraps a SyncHandlerAdapter with additional
// functionality. Middleware components can log, measure metrics, recover from
// panics, or modify the context before/after calling the next handler.
type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

// SyncHandlerAdapterFunc is an adapter type that allows regular functions
// to be used as SyncHandlerAdapter implementations.
type SyncHandlerAdapterFunc func(ctx context.Context, job bgjob.Job) Result

// Handle calls the underlying function.
func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, job bgjob.Job) Result {
	return a(ctx, job)
}

// Sync wraps a SyncHandlerAdapter with a chain of Middleware functions.
// It ensures that middleware is applied in the correct order (last to first).
type Sync struct {
	handler SyncHandlerAdapter
}

// NewSync creates a new Sync instance with the provided adapter and middleware chain.
// Middleware is applied in reverse order, so the first middleware in the list
// will be the outermost wrapper.
func NewSync(adapter SyncHandlerAdapter, middlewares ...Middleware) Sync {
	s := Sync{}
	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	s.handler = adapter
	return s
}

// Handle processes the job using the wrapped handler and middleware chain.
// It translates the Result into a bgjob.Result to control job lifecycle:
//   - Complete: marks the job as successfully processed
//   - Retry: schedules the job for retry after a delay
//   - MoveToDlq: moves the job to the dead letter queue
//   - Reschedule: reschedules the job for later execution
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
	case result.RescheduleWithArg:
		return bgjob.RescheduleWithArg(result.RescheduleDelay, result.Arg)
	}

	return bgjob.Result{}
}
