// Package consumer provides a high-level abstraction for consuming messages
// from Apache Kafka topics. It wraps the franz-go client and supports
// concurrent message processing with middleware chains for cross-cutting
// concerns like logging, metrics, and request IDs.
package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

// Middleware is a function that wraps a Handler to add cross-cutting
// functionality such as logging, metrics, or request ID propagation.
type Middleware func(next Handler) Handler

// Handler defines the interface for processing consumed messages.
type Handler interface {
	Handle(ctx context.Context, delivery *Delivery)
}

// HandlerFunc is an adapter that allows a function to be used as a Handler.
type HandlerFunc func(ctx context.Context, delivery *Delivery)

// Handle implements the Handler interface by calling the underlying function.
func (f HandlerFunc) Handle(ctx context.Context, delivery *Delivery) {
	f(ctx, delivery)
}

// Consumer handles consuming messages from Kafka topics with configurable
// concurrency and middleware support. It manages offset committing and
// provides lifecycle hooks through the observer pattern.
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

// New creates a new Consumer instance with the provided Kafka client, consumer
// group ID, handler, and concurrency level. Options can be used to configure
// middlewares and the observer.
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

// Run starts the consumer and begins processing messages. It returns
// immediately and runs message processing in a separate goroutine.
func (c *Consumer) Run(ctx context.Context) {
	if c.observer != nil {
		c.observer.BeginConsuming()
	}
	go c.run(ctx)
}

// Close gracefully shuts down the consumer, waits for pending message
// processing to complete, and releases the underlying Kafka client connection.
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

// Healthcheck returns nil if the consumer is healthy and able to fetch
// messages, or an error if it has encountered issues.
func (c *Consumer) Healthcheck(ctx context.Context) error {
	if c.alive.Load() {
		return nil
	}
	return errors.New("could not fetch messages")
}

// run is the main message processing loop. It polls for new messages and
// distributes them to worker goroutines for processing.
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

// handleFetchErrors handles errors that occur when polling for messages. It
// marks the consumer as unhealthy and notifies the observer.
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

// runWorker is a worker goroutine that processes messages from the deliveries
// channel. It continues until the channel is closed.
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

// handleMessage delegates message processing to the configured handler.
func (c *Consumer) handleMessage(ctx context.Context, delivery *Delivery) {
	c.handler.Handle(ctx, delivery)
}
