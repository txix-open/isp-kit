// Package bgjob_metrics provides Prometheus metric collectors for background job processing.
// It tracks job execution latency, success counts, retry counts, dead letter queue (DLQ) counts,
// and internal worker error counts.
//
// Example usage:
//
//	storage := bgjob_metrics.NewStorage(reg)
//	storage.ObserveExecuteDuration(queue, jobType, duration)
//	storage.IncSuccessCount(queue, jobType)
//	storage.IncInternalErrorCount()
package bgjob_metrics
