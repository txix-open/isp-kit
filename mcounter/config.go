package mcounter

import "time"

type CounterConfig struct {
	BufferCap     uint
	FlushInterval time.Duration
}

func DefaultConfig() *CounterConfig {
	return &CounterConfig{
		BufferCap:     10,
		FlushInterval: 5 * time.Second,
	}
}

func (c *CounterConfig) WithFlushInterval(interval time.Duration) *CounterConfig {
	c.FlushInterval = interval
	return c
}

func (c *CounterConfig) WithBufferCap(capacity uint) *CounterConfig {
	c.BufferCap = capacity
	return c
}
