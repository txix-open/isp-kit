package bgjobx

import (
	"time"

	"github.com/txix-open/isp-kit/bgjobx/handler"
)

type WorkerConfig struct {
	Queue        string
	Concurrency  int
	PollInterval time.Duration
	Handle       handler.SyncHandlerAdapter
}

func (c WorkerConfig) GetConcurrency() int {
	if c.Concurrency <= 0 {
		return 1
	}
	return c.Concurrency
}

func (c WorkerConfig) GetPollInterval() time.Duration {
	if c.PollInterval <= 0 {
		return 1 * time.Second
	}
	return c.PollInterval
}
