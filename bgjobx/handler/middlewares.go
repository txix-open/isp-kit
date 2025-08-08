package handler

import (
	"context"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

const (
	panicStackLength = 4 << 10
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
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) bgjob.Result {
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
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) (res bgjob.Result) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				var err error
				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = errors.Errorf("%v", recovered)
				}
				stack := make([]byte, panicStackLength)
				length := runtime.Stack(stack, false)
				res = bgjob.MoveToDlq(errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length]))
			}()
			return next.Handle(ctx, job)
		})
	}
}
