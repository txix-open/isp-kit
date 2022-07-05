package bgjobx

import (
	"context"
	"time"

	"github.com/integration-system/bgjob"
	"github.com/integration-system/isp-kit/log"
)

type LogObserver struct {
	log log.Logger
}

func (o LogObserver) JobStarted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job started", log.String("id", job.Id), log.String("type", job.Type), log.ByteString("arg", job.Arg))
}

func (o LogObserver) JobCompleted(ctx context.Context, job bgjob.Job) {
	o.log.Debug(ctx, "bgjob: job completed", log.String("id", job.Id))
}

func (o LogObserver) JobWillBeRetried(ctx context.Context, job bgjob.Job, after time.Duration, err error) {
	o.log.Error(ctx, "bgjob: job will be retried", log.String("id", job.Id), log.String("after", after.String()), log.Any("error", err))
}

func (o LogObserver) JobMovedToDlq(ctx context.Context, job bgjob.Job, err error) {
	o.log.Error(ctx, "bgjob: job will be moved to dlq", log.String("id", job.Id), log.Any("error", err))
}

func (o LogObserver) QueueIsEmpty(ctx context.Context) {
}

func (o LogObserver) WorkerError(ctx context.Context, err error) {
	o.log.Error(ctx, "bgjob: unexpected worker error", log.Any("error", err))
}
