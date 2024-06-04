package grmqx

import (
	"gitlab.txix.ru/isp/isp-kit/grmqx/handler"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	rabbitmq_metircs "gitlab.txix.ru/isp/isp-kit/metrics/rabbitmq_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/rabbitmq/consumer_tracing"
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
