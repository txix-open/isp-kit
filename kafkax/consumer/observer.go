package consumer

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

type Observer interface {
	ConsumerError(err error)
	BeginConsuming()
	CloseStart()
	CloseDone()
}

type NoopObserver struct{}

func (n NoopObserver) ConsumerError(err error) {}
func (n NoopObserver) BeginConsuming()         {}
func (n NoopObserver) CloseStart() {

}
func (n NoopObserver) CloseDone() {

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

func (l LogObserver) ConsumerError(err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected consumer error",
		log.Any("error", err),
	)
}

func (l LogObserver) BeginConsuming() {
	l.logger.Info(
		l.ctx,
		"kafka client: begin consuming",
	)
}

func (l LogObserver) CloseStart() {
	l.logger.Info(
		l.ctx,
		"kafka client: closing consumer start",
	)
}

func (l LogObserver) CloseDone() {
	l.logger.Info(
		l.ctx,
		"kafka client: closing consumer done",
	)
}
