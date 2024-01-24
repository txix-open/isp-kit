package worker

import (
	"time"
)

type Option func(w *Worker)

func WithInterval(interval time.Duration) Option {
	return func(w *Worker) {
		w.interval = interval
	}
}

func WithConcurrency(concurrency int) Option {
	return func(w *Worker) {
		w.concurrency = concurrency
	}
}
