package kafkax

import (
	"context"

	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
)

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

func (l LogObserver) ClientReady() {
	l.logger.Info(l.ctx, "kafka client: connected")
}

func (l LogObserver) ClientError(err error) {
	l.logger.Error(l.ctx, "kafka client: unexpected client error", log.Any("error", err))
}

func (l LogObserver) ConsumerError(consumer Consumer, err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected consumer error",
		log.String("topic", consumer.TopicName),
		log.Any("error", err),
	)
}

func (l LogObserver) ShutdownStarted() {
	l.logger.Info(l.ctx, "kafka client: shutdown was started")
}

func (l LogObserver) ShutdownDone() {
	l.logger.Info(l.ctx, "kafka client: shutdown was done")
}

func (l LogObserver) PublisherError(publisher *publisher.Publisher, err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected publisher error",
		log.String("topic", publisher.Topic),
		log.Any("error", err),
	)
}
