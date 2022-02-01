package grmqx

import (
	"context"

	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/isp-kit/log"
)

type LogObserver struct {
	ctx    context.Context
	logger log.Logger
}

func NewLogObserver(ctx context.Context, logger log.Logger) *LogObserver {
	return &LogObserver{
		ctx:    ctx,
		logger: logger,
	}
}

func (l LogObserver) ClientReady() {
	l.logger.Info(l.ctx, "rmq client: connected")
}

func (l LogObserver) ClientError(err error) {
	l.logger.Error(l.ctx, "rmq client: unexpected client error", log.Any("error", err))
}

func (l LogObserver) ConsumerError(consumer consumer.Consumer, err error) {
	l.logger.Error(
		l.ctx,
		"rmq client: unexpected consumer error",
		log.String("queue", consumer.Queue),
		log.Any("error", err),
	)
}

func (l LogObserver) ShutdownStarted() {
	l.logger.Info(l.ctx, "rmq client: shutdown was started")
}

func (l LogObserver) ShutdownDone() {
	l.logger.Info(l.ctx, "rmq client: shutdown was done")
}