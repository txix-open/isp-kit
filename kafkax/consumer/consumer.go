package consumer

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/atomic"
)

type Middleware func(next Handler) Handler

type Handler interface {
	Handle(ctx context.Context, delivery *Delivery)
}

type HandlerFunc func(ctx context.Context, delivery *Delivery)

func (f HandlerFunc) Handle(ctx context.Context, delivery *Delivery) {
	f(ctx, delivery)
}

type Consumer struct {
	reader *kafka.Reader

	middlewares []Middleware
	concurrency int
	handler     Handler
	observer    Observer
	deliveryWg  *sync.WaitGroup
	workersWg   *sync.WaitGroup
	deliveries  chan Delivery
	alive       *atomic.Bool
}

func New(reader *kafka.Reader, handler Handler, concurrency int, opts ...Option) *Consumer {
	if concurrency <= 0 {
		concurrency = 1
	}

	c := &Consumer{
		reader:      reader,
		concurrency: concurrency,
		handler:     handler,
		deliveryWg:  &sync.WaitGroup{},
		workersWg:   &sync.WaitGroup{},
		deliveries:  make(chan Delivery),
		alive:       atomic.NewBool(true),
	}

	for _, opt := range opts {
		opt(c)
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}
	c.handler = handler

	return c
}

func (c *Consumer) Run(ctx context.Context) {
	c.observer.BeginConsuming()
	go c.run(ctx)
}

func (c *Consumer) run(ctx context.Context) {
	for i := 0; i < c.concurrency; i++ {
		c.workersWg.Add(1)
		go c.runWorker(ctx)
	}

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			c.alive.Store(false)
			c.observer.ConsumerError(err)

			select {
			case <-ctx.Done():
			case <-time.After(1 * time.Second):
			}

			continue
		}

		c.alive.Store(true)
		delivery := NewDelivery(c.deliveryWg, c.reader, &msg, c.reader.Config().GroupID)
		c.deliveries <- *delivery
	}
}

//nolint:gosimple
func (c *Consumer) runWorker(ctx context.Context) {
	defer c.workersWg.Done()

	for {
		select {
		case delivery, isOpen := <-c.deliveries:
			if !isOpen { //normal close
				return
			}
			c.deliveryWg.Add(1)
			c.handleMessage(ctx, &delivery)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, delivery *Delivery) {
	c.handler.Handle(ctx, delivery)
}

func (c *Consumer) Close() error {
	defer func() {
		c.deliveryWg.Wait()
		close(c.deliveries)
		c.workersWg.Wait()
		c.alive.Store(false)

		c.observer.CloseDone()
	}()
	c.observer.CloseStart()

	err := c.reader.Close()
	if err != nil {
		return errors.WithMessage(err, "close kafka.reader")
	}

	return nil
}

func (c *Consumer) Healthcheck(ctx context.Context) error {
	if c.alive.Load() {
		return nil
	}
	return errors.New("could not fetch messages")
}
