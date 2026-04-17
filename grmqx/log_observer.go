package grmqx

import (
	"context"

	"github.com/txix-open/grmq"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/log"
)

// LogObserver implements grmq.Observer interface for logging RabbitMQ client events.
type LogObserver struct {
	grmq.NoopObserver

	ctx    context.Context
	logger log.Logger
}

// NewLogObserver creates a new log observer instance.
func NewLogObserver(ctx context.Context, logger log.Logger) LogObserver {
	return LogObserver{
		ctx:    ctx,
		logger: logger,
	}
}

// ClientReady logs a message indicating the RabbitMQ client is connected and ready.
func (l LogObserver) ClientReady() {
	l.logger.Info(l.ctx, "rmq client: connected")
}

// ClientError logs a message about an unexpected client error.
func (l LogObserver) ClientError(err error) {
	l.logger.Error(l.ctx, "rmq client: unexpected client error", log.Any("error", err))
}

// ConsumerError logs a message about an unexpected consumer error.
func (l LogObserver) ConsumerError(consumer consumer.Consumer, err error) {
	l.logger.Error(
		l.ctx,
		"rmq client: unexpected consumer error",
		log.String("queue", consumer.Queue),
		log.Any("error", err),
	)
}

// ShutdownStarted logs a message indicating the shutdown process has begun.
func (l LogObserver) ShutdownStarted() {
	l.logger.Info(l.ctx, "rmq client: shutdown was started")
}

// ShutdownDone logs a message indicating the shutdown process has completed.
func (l LogObserver) ShutdownDone() {
	l.logger.Info(l.ctx, "rmq client: shutdown was done")
}

// PublisherError logs a message about an unexpected publisher error.
func (l LogObserver) PublisherError(publisher *publisher.Publisher, err error) {
	l.logger.Error(
		l.ctx,
		"rmq client: unexpected publisher error",
		log.String("exchange", publisher.Exchange),
		log.String("routingKey", publisher.RoutingKey),
		log.Any("error", err),
	)
}

// PublishingFlow logs a message with information about the publishing flow state.
func (l LogObserver) PublishingFlow(publisher *publisher.Publisher, flow bool) {
	l.logger.Info(
		l.ctx,
		"rmq client: publishing flow",
		log.String("exchange", publisher.Exchange),
		log.String("routingKey", publisher.RoutingKey),
		log.Bool("running", flow),
	)
}

// ConnectionBlocked logs a message about the connection being blocked, including the reason.
func (l LogObserver) ConnectionBlocked(reason string) {
	l.logger.Error(
		l.ctx,
		"rmq client: connection blocked",
		log.String("reason", reason),
	)
}

// ConnectionUnblocked logs a message indicating the connection has been unblocked.
func (l LogObserver) ConnectionUnblocked() {
	l.logger.Info(l.ctx, "rmq client: connection unblocked")
}

// PublisherReconnected logs a message indicating the publisher has been reconnected.
func (l LogObserver) PublisherReconnected(publisher *publisher.Publisher) {
	l.logger.Info(
		l.ctx,
		"rmq client: publisher reconnected",
		log.String("exchange", publisher.Exchange),
		log.String("routingKey", publisher.RoutingKey),
	)
}
