package bgjobx

import (
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx/handler"
)

func NewDefaultHandler(adapter bgjob.Handler, metricStorage handler.MetricStorage) handler.Sync {
	return handler.NewSync(
		adapter,
		handler.WithDurationMeasure(metricStorage),
		handler.Recovery(),
	)
}
