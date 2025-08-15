package handler

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/panic_recovery"
)

type MetricStorage interface {
	ObserveExecuteDuration(queue string, jobType string, t time.Duration)
	IncRetryCount(queue string, jobType string)
	IncDlqCount(queue string, jobType string)
	IncSuccessCount(queue string, jobType string)
	IncInternalErrorCount()
}

func WithDurationMeasure(storage MetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) Result {
			start := time.Now()
			result := next.Handle(ctx, job)
			storage.ObserveExecuteDuration(job.Queue, job.Type, time.Since(start))
			return result
		})
	}
}

// nolint:nonamedreturns
func Recovery() Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) (res Result) {
			defer panic_recovery.Recover(func(err error) {
				res.MoveToDlq = true
				res.Err = err
			})
			return next.Handle(ctx, job)
		})
	}
}
