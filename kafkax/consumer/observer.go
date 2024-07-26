package consumer

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

type Observer interface {
	ConsumerError(consumer Consumer, err error)
}

type NoopObserver struct{}

func (n NoopObserver) ConsumerError(consumer Consumer, err error) {

}

type LogObserver struct {
	NoopObserver
	ctx    context.Context
	logger log.Logger
}

func NewLogObserver(ctx context.Context, logger log.Logger) LogObserver {
	return LogObserver{
		ctx:    ctx,
		logger: logger,
	}
}

func (l LogObserver) ConsumerError(consumer Consumer, err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected consumer error",
		log.String("topic", consumer.TopicName),
		log.Any("error", err),
	)
}
