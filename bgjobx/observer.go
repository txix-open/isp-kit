package bgjobx

import (
	"context"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

type Observer struct {
	log           log.Logger
	metricStorage handler.MetricStorage
}

func (o Observer) JobStarted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job started", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.String("type", job.Type))
}

func (o Observer) JobCompleted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job completed", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId))
	o.metricStorage.IncSuccessCount(job.Queue, job.Type)
}

// nolint: lll
func (o Observer) JobWillBeRetried(ctx context.Context, job bgjob.Job, after time.Duration, err error) {
	o.log.Error(ctx, "bgjob: job will be retried", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.String("after", after.String()), log.Any("error", err))
	o.metricStorage.IncRetryCount(job.Queue, job.Type)
}

func (o Observer) JobMovedToDlq(ctx context.Context, job bgjob.Job, err error) {
	o.log.Error(ctx, "bgjob: job will be moved to dlq", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.Any("error", err))
	o.metricStorage.IncDlqCount(job.Queue, job.Type)
}

// nolint: lll
func (o Observer) JobRescheduled(ctx context.Context, job bgjob.Job, after time.Duration) {
	o.log.Debug(ctx, "bgjob: job rescheduled", log.String("id", job.Id), log.String(requestid.LogKey, job.RequestId), log.Any("nextRunAt", after.String()))
}

func (o Observer) QueueIsEmpty(ctx context.Context) {
}

func (o Observer) WorkerError(ctx context.Context, err error) {
	o.log.Error(ctx, "bgjob: unexpected worker error", log.Any("error", err))
	o.metricStorage.IncInternalErrorCount()
}
