package grmqx

import (
	"github.com/integration-system/isp-kit/grmqx/handler"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/rabbitmq_metrics"
	"github.com/integration-system/isp-kit/observability/tracing/rabbitmq/consumer_tracing"
)

func NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync {
	return handler.NewSync(
		logger,
		adapter,
		handler.Log(logger),
		handler.Metrics(rabbitmq_metircs.NewConsumerStorage(metrics.DefaultRegistry)),
		consumer_tracing.NewConfig().Middleware(),
	)
}
