package bgjobx

import (
	"github.com/txix-open/isp-kit/bgjobx/handler"
)

// NewDefaultHandler creates a handler with standard middleware applied.
// It wraps the provided adapter with metrics collection, panic recovery,
// and request ID propagation middleware.
//
// The middleware stack is applied in the following order:
//  1. RequestId - propagates request IDs to the context
//  2. Recovery - catches panics and moves jobs to DLQ
//  3. Metrics - records execution duration and job outcomes
//
// Returns a Sync handler ready to be used with workers.
func NewDefaultHandler(adapter handler.SyncHandlerAdapter, metricStorage handler.MetricStorage) handler.Sync {
	return handler.NewSync(
		adapter,
		handler.Metrics(metricStorage),
		handler.Recovery(),
		handler.RequestId(),
	)
}
