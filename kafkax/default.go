package kafkax

import (
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
)

func NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync {
	return handler.NewSync(
		logger,
		adapter,
		handler.Log(logger),
		//handler.Metrics(rabbitmq_metircs.NewConsumerStorage(metrics.DefaultRegistry)),
		//consumer_tracing.NewConfig().Middleware(),
	)
}
