// Package batch_handler provides batch message processing for RabbitMQ consumers.
package batch_handler

import (
	"context"
	"sync"
	"time"

	"github.com/txix-open/grmq/consumer"
)

// Handler accumulates messages and processes them in batches.
// It triggers processing when the batch reaches maxSize or when the purgeInterval elapses.
type Handler struct {
	adapter       SyncHandlerAdapter
	purgeInterval time.Duration
	maxSize       int
	batch         []*BatchItem
	c             chan BatchItem
	runner        *sync.Once
	closed        bool
	lock          sync.Locker
}

// New creates a new batch handler with the specified adapter, purge interval, and max batch size.
func New(adapter SyncHandlerAdapter, purgeInterval time.Duration, maxSize int) *Handler {
	return &Handler{
		adapter:       adapter,
		purgeInterval: purgeInterval,
		maxSize:       maxSize,
		c:             make(chan BatchItem),
		runner:        &sync.Once{},
		lock:          &sync.Mutex{},
	}
}

// Handle adds a message to the batch and triggers processing if needed.
// If the handler is closed, the message is negatively acknowledged.
func (r *Handler) Handle(ctx context.Context, delivery *consumer.Delivery) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.closed {
		_ = delivery.Nack(true)
		return
	}

	r.runner.Do(func() {
		go r.run()
	})
	r.c <- BatchItem{
		Context:  ctx,
		Delivery: delivery,
	}
}

// Close stops the batch handler and prevents further message processing.
func (r *Handler) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.closed = true
	close(r.c)
}

// run is the main processing loop that accumulates messages and triggers batch processing.
func (r *Handler) run() {
	var timer *time.Timer
	defer func() {
		if len(r.batch) > 0 {
			r.adapter.Handle(r.batch)
		}
		if timer != nil {
			timer.Stop()
		}
	}()
	for {
		if timer == nil {
			timer = time.NewTimer(r.purgeInterval)
		} else {
			timer.Reset(r.purgeInterval)
		}

		select {
		case item, ok := <-r.c:
			if !ok {
				return
			}
			r.batch = append(r.batch, &item)
			if len(r.batch) < r.maxSize {
				continue
			}
		case <-timer.C:
			if len(r.batch) == 0 {
				continue
			}
		}

		r.adapter.Handle(r.batch)
		r.batch = nil
	}
}
