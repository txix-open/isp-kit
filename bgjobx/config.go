package bgjobx

import (
	"time"

	"github.com/txix-open/isp-kit/bgjobx/handler"
)

// WorkerConfig defines the configuration for a background job worker.
// It specifies the queue to poll, concurrency level, polling interval,
// and the handler that processes jobs.
type WorkerConfig struct {
	Queue        string
	Concurrency  int
	PollInterval time.Duration
	Handle       handler.SyncHandlerAdapter
}

// GetConcurrency returns the configured concurrency level.
// If Concurrency is not set or is less than or equal to zero, it returns 1.
func (c WorkerConfig) GetConcurrency() int {
	if c.Concurrency <= 0 {
		return 1
	}
	return c.Concurrency
}

// GetPollInterval returns the configured polling interval.
// If PollInterval is not set or is less than or equal to zero, it returns 1 second.
func (c WorkerConfig) GetPollInterval() time.Duration {
	if c.PollInterval <= 0 {
		return 1 * time.Second
	}
	return c.PollInterval
}
