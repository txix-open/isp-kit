package consumer

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

// Observer defines an interface for receiving lifecycle events from the
// consumer, such as error notifications and shutdown progress.
type Observer interface {
	ConsumerError(err error)
	BeginConsuming()
	CloseStart()
	CloseDone()
}

// NoopObserver is a no-op implementation of the Observer interface that
// ignores all events.
type NoopObserver struct{}

// ConsumerError does nothing.
func (n NoopObserver) ConsumerError(err error) {}

// BeginConsuming does nothing.
func (n NoopObserver) BeginConsuming() {}

// CloseStart does nothing.
func (n NoopObserver) CloseStart() {

}

// CloseDone does nothing.
func (n NoopObserver) CloseDone() {

}

// LogObserver is an Observer implementation that logs consumer lifecycle
// events to the provided logger.
type LogObserver struct {
	NoopObserver

	ctx    context.Context
	logger log.Logger
}

// NewLogObserver creates a new LogObserver with the provided context and logger.
func NewLogObserver(ctx context.Context, logger log.Logger) LogObserver {
	return LogObserver{
		ctx:    ctx,
		logger: logger,
	}
}

// ConsumerError logs an unexpected consumer error.
func (l LogObserver) ConsumerError(err error) {
	l.logger.Error(
		l.ctx,
		"kafka client: unexpected consumer error",
		log.Any("error", err),
	)
}

// BeginConsuming logs that the consumer has started processing messages.
func (l LogObserver) BeginConsuming() {
	l.logger.Info(
		l.ctx,
		"kafka client: begin consuming",
	)
}

// CloseStart logs that the consumer shutdown process has begun.
func (l LogObserver) CloseStart() {
	l.logger.Info(
		l.ctx,
		"kafka client: closing consumer start",
	)
}

// CloseDone logs that the consumer shutdown process has completed.
func (l LogObserver) CloseDone() {
	l.logger.Info(
		l.ctx,
		"kafka client: closing consumer done",
	)
}
