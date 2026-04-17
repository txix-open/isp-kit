package worker

import (
	"time"
)

// Option is a function type that configures a Worker.
type Option func(w *Worker)

// WithInterval sets the time interval between job executions.
// The default interval is 1 second if not specified.
//
// Example:
//
//	w := worker.New(myJob, worker.WithInterval(5*time.Second))
func WithInterval(interval time.Duration) Option {
	return func(w *Worker) {
		w.interval = interval
	}
}

// WithConcurrency sets the number of parallel workers.
// The default concurrency is 1 if not specified.
//
// Example:
//
//	w := worker.New(myJob, worker.WithConcurrency(3))
func WithConcurrency(concurrency int) Option {
	return func(w *Worker) {
		w.concurrency = concurrency
	}
}
