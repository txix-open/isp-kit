// Package publisher provides functionality for publishing messages to a STOMP broker.
package publisher

import (
	"context"
	"sync"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

// Middleware is a function that wraps a RoundTripper with additional functionality.
type Middleware func(next RoundTripper) RoundTripper

// RoundTripper defines the interface for publishing messages.
type RoundTripper interface {
	Publish(ctx context.Context, queue string, msg *Message) error
}

// RoundTripperFunc is an adapter type that allows using functions as round trippers.
type RoundTripperFunc func(ctx context.Context, queue string, msg *Message) error

// Publish calls the underlying function.
func (f RoundTripperFunc) Publish(ctx context.Context, queue string, msg *Message) error {
	return f(ctx, queue, msg)
}

// PublishOption is a function type for configuring message publication.
type PublishOption = func(*frame.Frame) error

// Publisher manages a connection to a STOMP broker for publishing messages.
type Publisher struct {
	// Address is the broker address.
	Address string
	// Queue is the default queue name to publish to.
	Queue string
	// ConnOpts are connection options.
	ConnOpts []consumer.ConnOption
	// Middlewares are middleware functions applied to the round tripper.
	Middlewares []Middleware

	roundTripper RoundTripper
	lock         sync.Locker
	conn         *stomp.Conn
}

// NewPublisher creates a new Publisher with the provided configuration options.
func NewPublisher(address string, queue string, opts ...Option) *Publisher {
	p := &Publisher{
		Address: address,
		Queue:   queue,
		lock:    &sync.Mutex{},
	}
	for _, opt := range opts {
		opt(p)
	}

	roundTripper := RoundTripper(RoundTripperFunc(p.publish))
	for i := len(p.Middlewares) - 1; i >= 0; i-- {
		roundTripper = p.Middlewares[i](roundTripper)
	}
	p.roundTripper = roundTripper

	return p
}

// Publish sends a message to the default queue configured for this publisher.
func (p *Publisher) Publish(ctx context.Context, msg *Message) error {
	return p.PublishTo(ctx, p.Queue, msg)
}

// PublishTo sends a message to the specified queue.
func (p *Publisher) PublishTo(ctx context.Context, queue string, msg *Message) error {
	return p.roundTripper.Publish(ctx, queue, msg)
}

// Close disconnects the publisher from the broker.
func (p *Publisher) Close() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.conn != nil {
		err := p.conn.Disconnect()
		p.conn = nil
		if err != nil {
			return errors.WithMessage(err, "disconnect")
		}
	}
	return nil
}

// Healthcheck verifies that the publisher can connect to the broker.
func (p *Publisher) Healthcheck(ctx context.Context) error {
	if len(p.Address) == 0 {
		return errors.New("publisher is not initialized")
	}

	_, err := p.aliveConn()
	if err != nil {
		return errors.WithMessage(err, "connect to stomp")
	}

	return nil
}

func (p *Publisher) publish(ctx context.Context, queue string, msg *Message) error {
	conn, err := p.aliveConn()
	if err != nil {
		return errors.WithMessage(err, "get alive connection")
	}
	err = conn.Send(queue, msg.ContentType, msg.Body, msg.Opts...)
	if err != nil {
		p.lock.Lock()
		_ = conn.MustDisconnect()
		p.conn = nil
		p.lock.Unlock()
		return errors.WithMessage(err, "stomp send")
	}
	return nil
}

func (p *Publisher) aliveConn() (*stomp.Conn, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.conn != nil {
		return p.conn, nil
	}

	conn, err := stomp.Dial("tcp", p.Address, p.ConnOpts...)
	if err != nil {
		return nil, errors.WithMessagef(err, "stomp dial to %s", p.Address)
	}
	p.conn = conn
	return conn, nil
}
