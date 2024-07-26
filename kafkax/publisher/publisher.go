package publisher

import (
	"context"
	"sync"

	"github.com/go-stomp/stomp/v3/frame"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/log"
	"go.uber.org/atomic"
)

type Middleware func(next RoundTripper) RoundTripper

type RoundTripper interface {
	Publish(ctx context.Context, msg *kafka.Message) error
}

type RoundTripperFunc func(ctx context.Context, msg *kafka.Message) error

func (f RoundTripperFunc) Publish(ctx context.Context, msg *kafka.Message) error {
	return f(ctx, msg)
}

type PublishOption = func(*frame.Frame) error

type Publisher struct {
	Topic       string
	Address     string
	Middlewares []Middleware

	logger       log.Logger
	observer     kafkax.Observer
	roundTripper RoundTripper
	lock         sync.Locker
	w            *kafka.Writer
	alive        *atomic.Bool
}

func New(writer *kafka.Writer, logger log.Logger, observer kafkax.Observer, opts ...Option) *Publisher {
	p := &Publisher{
		w:        writer,
		logger:   logger,
		alive:    atomic.NewBool(true),
		observer: observer,
		lock:     &sync.Mutex{},
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

func (p *Publisher) Publish(ctx context.Context, msg *kafka.Message) error {
	return p.PublishTo(ctx, msg)
}

func (p *Publisher) PublishTo(ctx context.Context, msg *kafka.Message) error {
	return p.roundTripper.Publish(ctx, msg)
}

func (p *Publisher) publish(ctx context.Context, msg *kafka.Message) error {
	err := p.w.WriteMessages(ctx, *msg)
	if err != nil {
		p.alive.Store(false)

		p.observer.PublisherError(p, err)

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
