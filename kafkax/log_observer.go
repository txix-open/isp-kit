package kafkax

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

// Observer defines an interface for receiving lifecycle events from the Kafka
// client, such as connection status and shutdown notifications.
type Observer interface {
	ClientReady()
	ClientError(err error)
	ShutdownStarted()
	ShutdownDone()
}

// NoopObserver is a no-op implementation of the Observer interface that
// ignores all events.
type NoopObserver struct{}

// ClientReady does nothing.
func (n NoopObserver) ClientReady() {

}

// ClientError does nothing.
func (n NoopObserver) ClientError(err error) {

}

// ShutdownStarted does nothing.
func (n NoopObserver) ShutdownStarted() {

}

// ShutdownDone does nothing.
func (n NoopObserver) ShutdownDone() {

}

// LogObserver is an Observer implementation that logs lifecycle events to the
// provided logger.
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

// ClientReady logs that the Kafka client has successfully connected.
func (l LogObserver) ClientReady() {
	l.logger.Info(l.ctx, "kafka client: connected")
}

// ClientError logs an unexpected client error.
func (l LogObserver) ClientError(err error) {
	l.logger.Error(l.ctx, "kafka client: unexpected client error", log.Any("error", err))
}

// ShutdownStarted logs that the shutdown process has begun.
func (l LogObserver) ShutdownStarted() {
	l.logger.Info(l.ctx, "kafka client: shutdown was started")
}

// ShutdownDone logs that the shutdown process has completed.
func (l LogObserver) ShutdownDone() {
	l.logger.Info(l.ctx, "kafka client: shutdown was done")
}
