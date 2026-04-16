package handler

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/panic_recovery"
	"github.com/txix-open/isp-kit/requestid"
)

// MetricStorage defines the interface for recording job execution metrics.
// Implementations should track job performance and outcome statistics.
type MetricStorage interface {
	// ObserveExecuteDuration records the time taken to execute a job.
	ObserveExecuteDuration(queue string, jobType string, t time.Duration)
	// IncRetryCount increments the retry counter for a job type.
	IncRetryCount(queue string, jobType string)
	// IncDlqCount increments the dead letter queue counter for a job type.
	IncDlqCount(queue string, jobType string)
	// IncSuccessCount increments the success counter for a job type.
	IncSuccessCount(queue string, jobType string)
	// IncInternalErrorCount increments the internal worker error counter.
	IncInternalErrorCount()
}

// Metrics creates a middleware that records execution metrics.
// It measures the duration of job execution and reports success, retry,
// and DLQ events to the provided MetricStorage.
func Metrics(storage MetricStorage) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) Result {
			start := time.Now()
			result := next.Handle(ctx, job)
			storage.ObserveExecuteDuration(job.Queue, job.Type, time.Since(start))
			return result
		})
	}
}

// Recovery creates a middleware that catches panics during job execution.
// When a panic occurs, the job is moved to the dead letter queue with
// the panic error. This prevents worker crashes from unhandled panics.
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

// RequestId creates a middleware that ensures request IDs are available
// in the handler context. If the job does not have a RequestId, it
// generates a new one. The request ID is added to the context for
// downstream logging and tracing.
func RequestId() Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) Result {
			requestId := job.RequestId

			if requestId == "" {
				requestId = requestid.Next()
				job.RequestId = requestId
			}

			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))

			return next.Handle(ctx, job)
		})
	}
}
