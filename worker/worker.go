// Package worker provides a configurable worker for periodic task execution
// with concurrency control and graceful shutdown support.
package worker

import (
	"context"
	"sync"
	"time"
)

// Job is an interface for tasks executed by the worker.
type Job interface {
	// Do executes the task.
	Do(ctx context.Context)
}

// Worker manages periodic execution of a Job with configurable concurrency.
type Worker struct {
	interval    time.Duration
	concurrency int
	wg          *sync.WaitGroup
	stop        chan struct{}
	job         Job
}

// New creates a new Worker with the specified Job and options.
// Default configuration: interval of 1 second and concurrency of 1.
//
// Example:
//
//	w := worker.New(myJob, worker.WithInterval(2*time.Second), worker.WithConcurrency(3))
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

// Run starts the worker goroutines. The operation is non-blocking.
// Each worker runs independently and processes jobs according to the configured interval.
//
// The context parameter controls the lifecycle of the workers. When the context is cancelled,
// workers will complete their current job and exit.
func (w *Worker) Run(ctx context.Context) {
	for range w.concurrency {
		w.wg.Add(1)
		go w.run(ctx)
	}
}

// Shutdown stops all workers and waits for them to complete.
// This method blocks until all goroutines have finished.
//
// Shutdown is safe for concurrent use.
func (w *Worker) Shutdown() {
	close(w.stop)
	w.wg.Wait()
}

// run is the main loop for a single worker goroutine.
// It executes the job, waits for the configured interval, and repeats until
// the context is cancelled or Shutdown is called.
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
