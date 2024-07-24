package kafkax

import (
	"context"

	"github.com/segmentio/kafka-go"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"go.uber.org/atomic"
)

const (
	bytesInMb        = 1024 * 1024
	defaultMsgSizeMb = 1
)

type PublisherMiddleware interface {
	Publish(ctx context.Context, msg kafka.Message) error
}

type Middleware func(next PublisherMiddleware) PublisherMiddleware

type PublisherMiddlewareFunc func(ctx context.Context, msg kafka.Message) error

func (f PublisherMiddlewareFunc) Publish(ctx context.Context, msg kafka.Message) error {
	return f(ctx, msg)
}

type Publisher struct {
	w        *kafka.Writer
	logger   log.Logger
	alive    *atomic.Bool
	connId   string
	Topic    string
	observer Observer
}

func (p *Publisher) Publish(ctx context.Context, msg kafka.Message) error {
	err := p.w.WriteMessages(ctx, msg)
	if err != nil {
		p.alive.Store(false)

		p.observer.PublisherError(p, err)

		return errors.WithMessage(err, "write messages")
	}
	p.alive.Store(true)

	p.logger.Info(ctx, "Kafka publish message",
		log.String("addr", p.w.Addr.String()),
		log.String("topic", p.w.Topic),
		log.String("connId", p.connId),
		log.ByteString("messageKey", msg.Key),
		log.ByteString("messageValue", msg.Value),
	)

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
