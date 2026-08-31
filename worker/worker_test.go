package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/worker"
)

type job struct {
	call atomic.Int32
}

func (j *job) Do(ctx context.Context) {
	j.call.Add(1)
}

func TestWorker_Run(t *testing.T) {
	t.Parallel()

	job := &job{}
	w := worker.New(job, worker.WithInterval(10*time.Second))
	w.Run(t.Context())
	time.Sleep(1 * time.Second)
	require.EqualValues(t, 1, job.call.Load())
	w.Shutdown()
}
