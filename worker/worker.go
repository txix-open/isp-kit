package worker

import (
	"context"
	"sync"
	"time"
)

type Job interface {
	Do(ctx context.Context)
}

type Worker struct {
	interval    time.Duration
	concurrency int
	wg          *sync.WaitGroup
	stop        chan struct{}
	job         Job
}

func New(job Job, opts ...Option) *Worker {
	w := &Worker{
		interval:    1 * time.Second,
		concurrency: 1,
		wg:          &sync.WaitGroup{},
		stop:        make(chan struct{}),
		job:         job,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *Worker) Run(ctx context.Context) {
	for range w.concurrency {
		w.wg.Add(1)
		go w.run(ctx)
	}
}

func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stop:
			return
		default:
		}

		w.job.Do(ctx)

		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-time.After(w.interval):
		}
	}
}

func (w *Worker) Shutdown() {
	close(w.stop)
	w.wg.Wait()
}
