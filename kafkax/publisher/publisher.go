// Package publisher provides a high-level abstraction for publishing messages
// to Apache Kafka topics. It wraps the franz-go client and supports middleware
// chains for cross-cutting concerns like metrics, logging, and request IDs.
package publisher

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

// Middleware is a function that wraps a RoundTripper to add cross-cutting
// functionality such as logging, metrics, or retry logic.
type Middleware func(next RoundTripper) RoundTripper

// RoundTripper defines the interface for publishing messages to Kafka.
type RoundTripper interface {
	Publish(ctx context.Context, rs ...*kgo.Record) error
}

// RoundTripperFunc is an adapter that allows a function to be used as a
// RoundTripper.
type RoundTripperFunc func(ctx context.Context, rs ...*kgo.Record) error

// Publish implements the RoundTripper interface by calling the underlying
// function.
func (f RoundTripperFunc) Publish(ctx context.Context, rs ...*kgo.Record) error {
	return f(ctx, rs...)
}

// Publisher handles publishing messages to a specific Kafka topic. It supports
// middleware chains for adding cross-cutting concerns and tracks the health
// status of the publisher.
type Publisher struct {
	client *kgo.Client

	topic       string
	middlewares []Middleware

	roundTripper RoundTripper
	lock         sync.Locker
	alive        *atomic.Bool
}

// New creates a new Publisher instance with the provided Kafka client and topic.
// Options can be used to configure middlewares and other settings.
func New(client *kgo.Client, topic string, opts ...Option) *Publisher {
	p := &Publisher{
		client: client,
		topic:  topic,
		alive:  atomic.NewBool(true),
		lock:   &sync.Mutex{},
	}

	for _, opt := range opts {
		opt(p)
	}

	roundTripper := RoundTripper(RoundTripperFunc(p.publish))
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		roundTripper = p.middlewares[i](roundTripper)
	}
	p.roundTripper = roundTripper

	return p
}

// Publish sends one or more messages to the configured topic. If a message's
// topic is not set, the publisher's default topic is used. Middlewares are
// applied in the order they were configured.
func (p *Publisher) Publish(ctx context.Context, rs ...*kgo.Record) error {
	for i, r := range rs {
		if len(r.Topic) == 0 {
			rs[i].Topic = p.topic
		}
	}
	return p.roundTripper.Publish(ctx, rs...)
}

// Close gracefully shuts down the publisher and releases the underlying Kafka
// client connection.
func (p *Publisher) Close() error {
	p.client.Close()
	return nil
}

// Healthcheck returns nil if the publisher is healthy, or an error if it has
// encountered issues during message production.
func (p *Publisher) Healthcheck(_ context.Context) error {
	if p.alive.Load() {
		return nil
	}
	return errors.New("kafka publisher: not healthy " + p.topic)
}

// publish is the core implementation that sends messages to Kafka using the
// synchronous produce API. It updates the health status based on the result.
func (p *Publisher) publish(ctx context.Context, rs ...*kgo.Record) error {
	results := p.client.ProduceSync(ctx, rs...)
	err := results.FirstErr()
	if err != nil {
		p.alive.Store(false)
		return errors.WithMessage(err, "write messages")
	}
	p.alive.Store(true)

	return nil
}
