package kafkax

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

type Observer interface {
	ClientReady()
	ClientError(err error)
	ShutdownStarted()
	ShutdownDone()
}

type NoopObserver struct{}

func (n NoopObserver) ClientReady() {

}

func (n NoopObserver) ClientError(err error) {

}

func (n NoopObserver) ShutdownStarted() {

}

func (n NoopObserver) ShutdownDone() {

}

// nolint:containedctx
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

func (l LogObserver) ShutdownStarted() {
	l.logger.Info(l.ctx, "kafka client: shutdown was started")
}

func (l LogObserver) ShutdownDone() {
	l.logger.Info(l.ctx, "kafka client: shutdown was done")
}
