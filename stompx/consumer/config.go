// Package consumer provides functionality for consuming messages from a STOMP broker.
package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
)

// Handler defines the interface for processing message deliveries.
type Handler interface {
	Handle(ctx context.Context, delivery *Delivery)
}

// HandlerFunc is an adapter type that allows using functions as handlers.
type HandlerFunc func(ctx context.Context, delivery *Delivery)

// Handle calls the underlying function.
func (f HandlerFunc) Handle(ctx context.Context, delivery *Delivery) {
	f(ctx, delivery)
}

// Middleware is a function that wraps a Handler with additional functionality.
type Middleware func(next Handler) Handler

// ConnOption is a function type for configuring STOMP connections.
type ConnOption = func(*stomp.Conn) error

// SubscriptionOption is a function type for configuring subscriptions.
type SubscriptionOption = func(*frame.Frame) error

// Config holds the configuration for a message consumer.
type Config struct {
	// Address is the broker address.
	Address string
	// Queue is the queue name to consume from.
	Queue string
	// ConnOpts are connection options.
	ConnOpts []ConnOption
	// Concurrency is the number of concurrent workers.
	Concurrency int
	// Middlewares are middleware functions applied to the handler.
	Middlewares []Middleware
	// SubscriptionOpts are subscription options.
	SubscriptionOpts []SubscriptionOption
	// Observer is the lifecycle observer.
	Observer Observer
	// ReconnectTimeout is the delay before reconnecting.
	ReconnectTimeout time.Duration

	handler Handler
}

// NewConfig creates a new consumer configuration with the provided options.
func NewConfig(address string, queue string, handler Handler, opts ...Option) Config {
	c := &Config{
		Address:          address,
		Queue:            queue,
		Concurrency:      1,
		ReconnectTimeout: 1 * time.Second,
		Observer:         NoopObserver{},
	}
	for _, opt := range opts {
		opt(c)
	}

	for i := len(c.Middlewares) - 1; i >= 0; i-- {
		handler = c.Middlewares[i](handler)
	}
	c.handler = handler

	return *c
}

// String returns a string representation of the configuration.
func (c Config) String() string {
	return fmt.Sprintf("%s/%s", c.Address, c.Queue)
}
