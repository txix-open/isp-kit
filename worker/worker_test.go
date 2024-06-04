package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.txix.ru/isp/isp-kit/worker"
)

type job struct {
	call int32
}

func (j *job) Do(ctx context.Context) {
	atomic.AddInt32(&j.call, 1)
}

func TestWorker_Run(t *testing.T) {
	job := &job{}
	w := worker.New(job, worker.WithInterval(10*time.Second))
	w.Run(context.Background())
	time.Sleep(1 * time.Second)
	require.EqualValues(t, 1, atomic.LoadInt32(&job.call))
	w.Shutdown()
}
