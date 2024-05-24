package grmqx

import (
	"context"
	"sync"
	"time"

	"github.com/txix-open/grmq/consumer"
)

type BatchHandlerAdapter interface {
	Handle(batch []BatchItem)
}

type BatchHandlerAdapterFunc func(batch []BatchItem)

func (b BatchHandlerAdapterFunc) Handle(batch []BatchItem) {
	b(batch)
}

type BatchItem struct {
	Context  context.Context
	Delivery *consumer.Delivery
}

type BatchHandler struct {
	adapter       BatchHandlerAdapter
	purgeInterval time.Duration
	maxSize       int
	batch         [][]BatchItem
	c             chan BatchItem
	runner        []*sync.Once
	closed        bool
	lock          sync.Locker
	concurrency   int
}

type BatchHandlerOption func(h *BatchHandler)

func WithConcurrency(concurrency int) BatchHandlerOption {
	return func(bh *BatchHandler) {
		fillConcurrency(bh, concurrency)
	}
}

func fillConcurrency(bh *BatchHandler, concurrency int) {
	bh.concurrency = concurrency
	bh.runner = make([]*sync.Once, concurrency)
	for i := range concurrency {
		bh.runner[i] = &sync.Once{}
	}
	bh.batch = make([][]BatchItem, concurrency)
}

func NewBatchHandler(
	adapter BatchHandlerAdapter,
	purgeInterval time.Duration,
	maxSize int,
	opts ...BatchHandlerOption,
) *BatchHandler {
	bh := BatchHandler{
		adapter:       adapter,
		purgeInterval: purgeInterval,
		maxSize:       maxSize,
		c:             make(chan BatchItem),
		lock:          &sync.Mutex{},
		concurrency:   1,
	}

	for _, opt := range opts {
		opt(&bh)
	}

	if bh.concurrency == 1 {
		fillConcurrency(&bh, 1)
	}

	return &bh
}

func (r *BatchHandler) Handle(ctx context.Context, delivery *consumer.Delivery) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.closed {
		_ = delivery.Nack(true)
	}

	for idx := range r.concurrency {
		r.runner[idx].Do(func() {
			go r.run(idx)
		})
	}

	r.c <- BatchItem{
		Context:  ctx,
		Delivery: delivery,
	}
}

func (r *BatchHandler) run(idx int) {
	var timer *time.Timer
	defer func() {
		if len(r.batch[idx]) > 0 {
			r.adapter.Handle(r.batch[idx])
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
			r.batch[idx] = append(r.batch[idx], item)
			if len(r.batch[idx]) < r.maxSize {
				continue
			}
		case <-timer.C:
			if len(r.batch[idx]) == 0 {
				continue
			}
		}

		r.adapter.Handle(r.batch[idx])
		r.batch[idx] = make([]BatchItem, 0)
	}
}

func (r *BatchHandler) Close() {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.closed = true
	close(r.c)
}
