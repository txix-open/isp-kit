package metrics

import (
	"time"
)

// nolint:gochecknoglobals,mnd
var (
	// DefaultObjectives defines the quantile targets for Prometheus summaries.
	// Each key represents a quantile (e.g., 0.95 for the 95th percentile),
	// and each value represents the allowed error tolerance.
	DefaultObjectives = map[float64]float64{
		0.5:  0.05,  // 0.45-0.55
		0.9:  0.01,  // 0.89-0.91
		0.95: 0.005, // 0.94.5-0.95.5
		0.99: 0.001, // 0.98.9-0.99.1,
	}
)

// Milliseconds converts a time.Duration to milliseconds as a float64.
// Returns 1 if the duration is zero
func Milliseconds(duration time.Duration) float64 {
	value := duration.Milliseconds()
	if value == 0 {
		return 1
	}
	return float64(value)
}
