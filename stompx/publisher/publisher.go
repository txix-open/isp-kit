package publisher

import (
	"context"
	"sync"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/integration-system/isp-kit/stompx/consumer"
	"github.com/pkg/errors"
)

type Middleware func(next RoundTripper) RoundTripper

type RoundTripper interface {
	Publish(ctx context.Context, queue string, msg *Message) error
}

type RoundTripperFunc func(ctx context.Context, queue string, msg *Message) error

func (f RoundTripperFunc) Publish(ctx context.Context, queue string, msg *Message) error {
	return f(ctx, queue, msg)
}

type PublishOption = func(*frame.Frame) error

type Publisher struct {
	Address     string
	Queue       string
	ConnOpts    []consumer.ConnOption
	Middlewares []Middleware

	roundTripper RoundTripper
	lock         sync.Locker
	conn         *stomp.Conn
}

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

func (p *Publisher) Publish(ctx context.Context, msg *Message) error {
	return p.PublishTo(ctx, p.Queue, msg)
}

func (p *Publisher) PublishTo(ctx context.Context, queue string, msg *Message) error {
	return p.roundTripper.Publish(ctx, queue, msg)
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
