package bgjobx

import (
	"github.com/txix-open/isp-kit/bgjobx/handler"
)

func NewDefaultHandler(adapter handler.SyncHandlerAdapter, metricStorage handler.MetricStorage) handler.Sync {
	return handler.NewSync(
		adapter,
		handler.Metrics(metricStorage),
		handler.Recovery(),
		handler.RequestId(),
	)
}
