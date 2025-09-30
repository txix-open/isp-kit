package batch_handler

import (
	"context"
	"sync"
	"time"

	"github.com/txix-open/grmq/consumer"
)

type BatchHandlerAdapter interface {
	Handle(batch []Item)
}

type BatchHandlerAdapterFunc func(batch []Item)

func (b BatchHandlerAdapterFunc) Handle(batch []Item) {
	b(batch)
}

// nolint:containedctx
type Item struct {
	Context  context.Context
	Delivery *consumer.Delivery
}

type Handler struct {
	adapter       BatchHandlerAdapter
	purgeInterval time.Duration
	maxSize       int
	batch         []Item
	c             chan Item
	runner        *sync.Once
	closed        bool
	lock          sync.Locker
}

func New(adapter BatchHandlerAdapter, purgeInterval time.Duration, maxSize int) *Handler {
	return &Handler{
		adapter:       adapter,
		purgeInterval: purgeInterval,
		maxSize:       maxSize,
		c:             make(chan Item),
		runner:        &sync.Once{},
		lock:          &sync.Mutex{},
	}
}

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
	r.c <- Item{
		Context:  ctx,
		Delivery: delivery,
	}
}

func (r *Handler) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.closed = true
	close(r.c)
}

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
			r.batch = append(r.batch, item)
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
