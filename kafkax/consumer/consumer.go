package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/pkg/errors"
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
	client *kgo.Client

	consumerGroupId string
	middlewares     []Middleware
	concurrency     int
	handler         Handler
	observer        Observer

	deliveryWg *sync.WaitGroup
	deliveries chan Delivery
	alive      *atomic.Bool

	stopChan chan struct{}
}

func New(client *kgo.Client, consumerGroupId string, handler Handler, concurrency int, opts ...Option) *Consumer {
	if concurrency <= 0 {
		concurrency = 1
	}

	c := &Consumer{
		client:          client,
		consumerGroupId: consumerGroupId,
		concurrency:     concurrency,
		handler:         handler,
		deliveryWg:      &sync.WaitGroup{},
		deliveries:      make(chan Delivery),
		alive:           atomic.NewBool(true),
		stopChan:        make(chan struct{}),
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
	if c.observer != nil {
		c.observer.BeginConsuming()
	}
	go c.run(ctx)
}

func (c *Consumer) Close() error {
	defer func() {
		c.alive.Store(false)
		if c.observer != nil {
			c.observer.CloseDone()
		}
	}()
	if c.observer != nil {
		c.observer.CloseStart()
	}
	close(c.stopChan)

	c.deliveryWg.Wait()

	c.client.Close()

	return nil
}

func (c *Consumer) Healthcheck(ctx context.Context) error {
	if c.alive.Load() {
		return nil
	}
	return errors.New("could not fetch messages")
}

func (c *Consumer) run(ctx context.Context) {
	for range c.concurrency {
		go c.runWorker(ctx)
	}

	defer close(c.deliveries)

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			c.handleFetchErrors(ctx, errs)
			continue
		}

		c.alive.Store(true)

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, msg := range p.Records {
				c.deliveryWg.Add(1)
				select {
				case <-c.stopChan:
					c.deliveryWg.Done()
					return
				default:
					delivery := NewDelivery(c.deliveryWg, c.client, msg, c.consumerGroupId)
					c.deliveries <- *delivery
				}
			}
		})
	}
}

func (c *Consumer) handleFetchErrors(ctx context.Context, errs []kgo.FetchError) {
	c.alive.Store(false)
	if c.observer != nil {
		for _, err := range errs {
			c.observer.ConsumerError(err.Err)
		}
	}

	select {
	case <-ctx.Done():
	case <-time.After(1 * time.Second):
	}
}

//nolint:gosimple
func (c *Consumer) runWorker(ctx context.Context) {
	for {
		select {
		case delivery, isOpen := <-c.deliveries:
			if !isOpen { // normal close
				return
			}

			c.handleMessage(ctx, &delivery)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, delivery *Delivery) {
	c.handler.Handle(ctx, delivery)
}
