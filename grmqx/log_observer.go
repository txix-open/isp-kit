package grmqx

import (
	"context"

	"github.com/txix-open/grmq"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/log"
)

// nolint:containedctx
type LogObserver struct {
	grmq.NoopObserver
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

func (l LogObserver) PublisherError(publisher *publisher.Publisher, err error) {
	l.logger.Error(
		l.ctx,
		"rmq client: unexpected publisher error",
		log.String("exchange", publisher.Exchange),
		log.String("routingKey", publisher.RoutingKey),
		log.Any("error", err),
	)
}

func (l LogObserver) PublishingFlow(publisher *publisher.Publisher, flow bool) {
	l.logger.Info(
		l.ctx,
		"rmq client: publishing flow",
		log.String("exchange", publisher.Exchange),
		log.String("routingKey", publisher.RoutingKey),
		log.Bool("running", flow),
	)
}

func (l LogObserver) ConnectionBlocked(reason string) {
	l.logger.Error(
		l.ctx,
		"rmq client: connection blocked",
		log.String("reason", reason),
	)
}

func (l LogObserver) ConnectionUnblocked() {
	l.logger.Info(l.ctx, "rmq client: connection unblocked")
}
