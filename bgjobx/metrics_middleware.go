package bgjobx

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
)

type MetricStorage interface {
	ObserveExecuteDuration(queue string, jobType string, t time.Duration)
	IncRetryCount(queue string, jobType string)
	IncDlqCount(queue string, jobType string)
	IncSuccessCount(queue string, jobType string)
	IncInternalErrorCount()
}

// nolint:ireturn
func WithDurationMeasure(storage MetricStorage, handler bgjob.Handler) bgjob.Handler {
	return bgjob.HandlerFunc(func(ctx context.Context, job bgjob.Job) bgjob.Result {
		start := time.Now()
		result := handler.Handle(ctx, job)
		storage.ObserveExecuteDuration(job.Queue, job.Type, time.Since(start))
		return result
	})
}
