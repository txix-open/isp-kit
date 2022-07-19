package bgjobx

import (
	"context"
	"time"

	"github.com/integration-system/bgjob"
	"github.com/integration-system/isp-kit/log"
)

type Observer struct {
	log           log.Logger
	metricStorage MetricStorage
}

func (o Observer) JobStarted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job started", log.String("id", job.Id), log.String("type", job.Type))
}

func (o Observer) JobCompleted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job completed", log.String("id", job.Id))
	o.metricStorage.IncSuccessCount(job.Queue, job.Type)
}

func (o Observer) JobWillBeRetried(ctx context.Context, job bgjob.Job, after time.Duration, err error) {
	o.log.Error(ctx, "bgjob: job will be retried", log.String("id", job.Id), log.String("after", after.String()), log.Any("error", err))
	o.metricStorage.IncRetryCount(job.Queue, job.Type)
}

func (o Observer) JobMovedToDlq(ctx context.Context, job bgjob.Job, err error) {
	o.log.Error(ctx, "bgjob: job will be moved to dlq", log.String("id", job.Id), log.Any("error", err))
	o.metricStorage.IncDlqCount(job.Queue, job.Type)
}

func (o Observer) QueueIsEmpty(ctx context.Context) {
}

func (o Observer) WorkerError(ctx context.Context, err error) {
	o.log.Error(ctx, "bgjob: unexpected worker error", log.Any("error", err))
	o.metricStorage.IncInternalErrorCount()
}
