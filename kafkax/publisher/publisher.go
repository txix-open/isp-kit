package publisher

import (
	"context"
	"sync"

	"github.com/go-stomp/stomp/v3/frame"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
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

type PublishOption = func(*frame.Frame) error

type Publisher struct {
	Topic       string
	Address     string
	Middlewares []Middleware

	logger       log.Logger
	observer     Observer
	roundTripper RoundTripper
	lock         sync.Locker
	w            *kafka.Writer
	alive        *atomic.Bool
}

func New(writer *kafka.Writer, topic string, logger log.Logger, opts ...Option) *Publisher {
	p := &Publisher{
		Topic:   topic,
		Address: writer.Addr.String(),
		w:       writer,
		logger:  logger,
		alive:   atomic.NewBool(true),
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

func (p *Publisher) Publish(ctx context.Context, msgs ...kafka.Message) error {
	for i, msg := range msgs {
		if len(msg.Topic) == 0 {
			msgs[i].Topic = p.Topic
		}
	}
	return p.PublishTo(ctx, msgs...)
}

func (p *Publisher) PublishTo(ctx context.Context, msgs ...kafka.Message) error {
	return p.roundTripper.Publish(ctx, msgs...)
}

func (p *Publisher) publish(ctx context.Context, msgs ...kafka.Message) error {
	err := p.w.WriteMessages(ctx, msgs...)
	if err != nil {
		p.alive.Store(false)
		p.observer.PublisherError(err)
		return errors.WithMessage(err, "write messages")
	}
	p.alive.Store(true)

	return nil
}

func (p *Publisher) Close() error {
	err := p.w.Close()
	if err != nil {
		return errors.WithMessage(err, "close writer")
	}

	return nil
}

func (p *Publisher) Healthcheck(_ context.Context) error {
	if p.alive.Load() {
		return nil
	}
	return errors.New("kafka publisher: not healthy " + p.w.Topic)
}
