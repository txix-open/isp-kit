package publisher

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/atomic"
)

type Middleware func(next RoundTripper) RoundTripper

type RoundTripper interface {
	Publish(ctx context.Context, msgs ...kafka.Message) error
}

type RoundTripperFunc func(ctx context.Context, msgs ...kafka.Message) error

func (f RoundTripperFunc) Publish(ctx context.Context, msgs ...kafka.Message) error {
	return f(ctx, msgs...)
}

type Publisher struct {
	Writer *kafka.Writer

	topic       string
	address     string
	middlewares []Middleware

	roundTripper RoundTripper
	lock         sync.Locker
	alive        *atomic.Bool
	metrics      *Metrics
}

func New(writer *kafka.Writer, topic string, metrics *Metrics, opts ...Option) *Publisher {
	p := &Publisher{
		Writer:  writer,
		topic:   topic,
		address: writer.Addr.String(),
		alive:   atomic.NewBool(true),
		lock:    &sync.Mutex{},
		metrics: metrics,
	}

	for _, opt := range opts {
		opt(p)
	}

	roundTripper := RoundTripper(RoundTripperFunc(p.publish))
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		roundTripper = p.middlewares[i](roundTripper)
	}
	p.roundTripper = roundTripper

	if p.metrics != nil {
		go p.metrics.Run()
	}

	return p
}

func (p *Publisher) Publish(ctx context.Context, msgs ...kafka.Message) error {
	for i, msg := range msgs {
		if len(msg.Topic) == 0 {
			msgs[i].Topic = p.topic
		}
	}
	return p.roundTripper.Publish(ctx, msgs...)
}

func (p *Publisher) publish(ctx context.Context, msgs ...kafka.Message) error {
	err := p.Writer.WriteMessages(ctx, msgs...)
	if err != nil {
		p.alive.Store(false)
		return errors.WithMessage(err, "write messages")
	}
	p.alive.Store(true)

	return nil
}

func (p *Publisher) Close() error {
	defer func() {
		if p.metrics != nil {
			p.metrics.Close()
		}
	}()
	err := p.Writer.Close()
	if err != nil {
		return errors.WithMessage(err, "close writer")
	}

	return nil
}

func (p *Publisher) Healthcheck(_ context.Context) error {
	if p.alive.Load() {
		return nil
	}
	return errors.New("kafka publisher: not healthy " + p.Writer.Topic)
}
