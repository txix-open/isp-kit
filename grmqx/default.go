package grmqx

import (
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	"github.com/txix-open/isp-kit/grmqx/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/rabbitmq_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/rabbitmq/consumer_tracing"
)

func NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync {
	return handler.NewSync(
		logger,
		adapter,
		handler.Log(logger),
		handler.Metrics(rabbitmq_metrics.NewConsumerStorage(metrics.DefaultRegistry)),
		consumer_tracing.NewConfig().Middleware(),
		handler.Recovery(),
	)
}

func NewResultBatchHandler(logger log.Logger, adapter batch_handler.SyncHandlerAdapter) batch_handler.Sync {
	return batch_handler.NewSync(
		logger,
		adapter,
		batch_handler.Log(logger),
		batch_handler.Metrics(rabbitmq_metrics.NewConsumerStorage(metrics.DefaultRegistry)),
		batch_handler.Recovery(logger),
	)
}
