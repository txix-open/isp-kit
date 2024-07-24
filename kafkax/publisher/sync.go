package publisher

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
)

type SyncPublisherAdapter interface {
	Publish(ctx context.Context, msg *kafka.Message) error
}

type Middleware func(next SyncPublisherAdapter) SyncPublisherAdapter

type SyncPublisherAdapterFunc func(ctx context.Context, msg *kafka.Message) error

func (f SyncPublisherAdapterFunc) Publish(ctx context.Context, msg *kafka.Message) error {
	return f(ctx, msg)
}

type Sync struct {
	logger    log.Logger
	publisher SyncPublisherAdapter
}

func NewSync(logger log.Logger, publisher SyncPublisherAdapter, middlewares ...Middleware) Sync {
	s := Sync{
		logger: logger,
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		publisher = middlewares[i](publisher)
	}
	s.publisher = publisher

	return s
}

func (r Sync) Publish(ctx context.Context, msg *kafka.Message) error {
	return r.publisher.Publish(ctx, msg)
}
