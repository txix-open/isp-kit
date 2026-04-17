package bgjobx

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

// Observer implements the bgjob.Observer interface to provide logging and
// metrics collection for job lifecycle events. It records job state changes
// such as completion, retries, and failures to the dead letter queue.
type Observer struct {
	bgjob.NoopObserver

	log           log.Logger
	metricStorage handler.MetricStorage
}

// JobStarted is called when a job begins processing.
// It logs the job ID, request ID, and job type at debug level.
func (o Observer) JobStarted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job started", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.String("type", job.Type))
}

// JobCompleted is called when a job finishes successfully.
// It logs the completion and increments the success counter for the queue and job type.
func (o Observer) JobCompleted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job completed", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId))
	o.metricStorage.IncSuccessCount(job.Queue, job.Type)
}

// JobWillBeRetried is called when a job fails and is scheduled for retry.
// It logs the error and the retry delay, and increments the retry counter.
func (o Observer) JobWillBeRetried(ctx context.Context, job bgjob.Job, after time.Duration, err error) {
	o.log.Error(ctx, "bgjob: job will be retried", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.String("after", after.String()), log.Any("error", err))
	o.metricStorage.IncRetryCount(job.Queue, job.Type)
}

// JobMovedToDlq is called when a job is moved to the dead letter queue
// after exhausting all retry attempts. It logs the error and increments the DLQ counter.
func (o Observer) JobMovedToDlq(ctx context.Context, job bgjob.Job, err error) {
	o.log.Error(ctx, "bgjob: job will be moved to dlq", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.Any("error", err))
	o.metricStorage.IncDlqCount(job.Queue, job.Type)
}

// JobRescheduled is called when a job is rescheduled for future execution.
// It logs the rescheduled time for the job.
func (o Observer) JobRescheduled(ctx context.Context, job bgjob.Job, after time.Duration) {
	o.log.Debug(ctx, "bgjob: job rescheduled", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.Any("nextRunAt", after.String()))
}

// WorkerError is called when an unexpected error occurs in the worker.
// It logs the error and increments the internal error counter.
func (o Observer) WorkerError(ctx context.Context, err error) {
	o.log.Error(ctx, "bgjob: unexpected worker error", log.Any("error", err))
	o.metricStorage.IncInternalErrorCount()
}
