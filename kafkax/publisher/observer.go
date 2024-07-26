package publisher

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

type Observer interface {
	PublisherError(err error)
}

type NoopObserver struct {
}

func (n NoopObserver) PublisherError(err error) {
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

func (l LogObserver) PublisherError(err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected publisher error",
		log.Any("error", err),
	)
}
