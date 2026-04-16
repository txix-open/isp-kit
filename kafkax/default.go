package kafkax

import (
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/kafka_metrics"
)

// NewResultHandler creates a new synchronous message handler with default
// middlewares including logging, metrics, and panic recovery. The provided
// adapter implements the business logic for handling messages.
func NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync {
	return handler.NewSync(
		logger,
		adapter,
		handler.Log(logger),
		handler.Metrics(kafka_metrics.NewConsumerStorage(metrics.DefaultRegistry)),
		handler.Recovery(),
	)
}
