package kafkax

import (
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/kafkax/publisher"
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

func NewPublisher(logger log.Logger, pub publisher.SyncPublisherAdapter, w *kafka.Writer, connId string) publisher.Sync {
	return publisher.NewSync(
		logger,
		pub,
		publisher.PublisherLog(logger, w, connId),
	)
}
