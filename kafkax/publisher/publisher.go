package publisher

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/stats"
	"github.com/txix-open/isp-kit/metrics"
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

	roundTripper   RoundTripper
	lock           sync.Locker
	alive          *atomic.Bool
	metricTimer    *time.Ticker
	stopMetricChan chan struct{}
	metricStorage  MetricStorage
}

func New(writer *kafka.Writer, topic string, sendMetricPeriod time.Duration, opts ...Option) *Publisher {
	p := &Publisher{
		Writer:         writer,
		topic:          topic,
		address:        writer.Addr.String(),
		alive:          atomic.NewBool(true),
		lock:           &sync.Mutex{},
		metricTimer:    time.NewTicker(sendMetricPeriod),
		stopMetricChan: make(chan struct{}),
		metricStorage:  stats.NewPublisherStorage(metrics.DefaultRegistry),
	}

	for _, opt := range opts {
		opt(p)
	}

	roundTripper := RoundTripper(RoundTripperFunc(p.publish))
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		roundTripper = p.middlewares[i](roundTripper)
	}
	p.roundTripper = roundTripper

	go p.runMetricSender()
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

func (p *Publisher) runMetricSender() {
	for {
		select {
		case _, isOpen := <-p.stopMetricChan:
			if !isOpen {
				return
			}

			return
		case <-p.metricTimer.C:
			p.sendMetrics(p.Writer.Stats())
		}
	}
}

func (p *Publisher) Close() error {
	defer func() {
		p.stopMetricChan <- struct{}{}
		close(p.stopMetricChan)
	}()

	p.metricTimer.Stop()
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
