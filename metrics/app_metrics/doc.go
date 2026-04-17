// Package app_metrics provides application-level metrics including log sampling statistics.
// It tracks the count of sampled and dropped log entries at different log levels.
//
// Example usage:
//
//	logCounter := app_metrics.NewLogCounter(reg)
//	logger := zap.New(core,
//		zap.Hooks(
//			logCounter.SampledLogCounter(),
//			logCounter.DroppedLogCounter(),
//		),
//	)
package app_metrics
