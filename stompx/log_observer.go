package stompx

import (
	"context"

	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

// LogObserver observes events in the consumer lifecycle (errors, startup, shutdown).
type LogObserver struct {
	logger log.Logger
}

// NewLogObserver creates an observer for lifecycle events with the specified logger.
func NewLogObserver(logger log.Logger) LogObserver {
	return LogObserver{logger: logger}
}

// Error logs consumer errors.
func (l LogObserver) Error(c *consumer.Consumer, err error) {
	l.logger.Error(
		context.Background(),
		"stomp client: unexpected consumer error",
		log.String("consumer", c.String()),
		log.Any("error", err),
	)
}

// CloseStart logs the start of the shutdown process.
func (l LogObserver) CloseStart(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: closing consumer start",
		log.String("consumer", c.String()),
	)
}

// CloseDone logs the completion of the shutdown process.
func (l LogObserver) CloseDone(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: closing consumer done",
		log.String("consumer", c.String()),
	)
}

// BeginConsuming logs the start of message consumption.
func (l LogObserver) BeginConsuming(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: begin consuming",
		log.String("consumer", c.String()),
	)
}
