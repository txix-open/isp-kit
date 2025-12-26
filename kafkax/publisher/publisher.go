package publisher

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/atomic"
)

type Middleware func(next RoundTripper) RoundTripper

type RoundTripper interface {
	Publish(ctx context.Context, rs ...*kgo.Record) error
}

type RoundTripperFunc func(ctx context.Context, rs ...*kgo.Record) error

func (f RoundTripperFunc) Publish(ctx context.Context, rs ...*kgo.Record) error {
	return f(ctx, rs...)
}

type Publisher struct {
	client *kgo.Client

	topic       string
	middlewares []Middleware

	roundTripper RoundTripper
	lock         sync.Locker
	alive        *atomic.Bool
}

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

func (p *Publisher) Publish(ctx context.Context, rs ...*kgo.Record) error {
	for i, r := range rs {
		if len(r.Topic) == 0 {
			rs[i].Topic = p.topic
		}
	}
	return p.roundTripper.Publish(ctx, rs...)
}

func (p *Publisher) Close() error {
	p.client.Close()
	return nil
}

func (p *Publisher) Healthcheck(_ context.Context) error {
	if p.alive.Load() {
		return nil
	}
	return errors.New("kafka publisher: not healthy " + p.topic)
}

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
