package metrics

import (
	"time"
)

// nolint:gochecknoglobals,mnd
var (
	DefaultObjectives = map[float64]float64{
		0.5:  0.05,  // 0.45-0.55
		0.9:  0.01,  // 0.89-0.91
		0.95: 0.005, // 0.94.5-0.95.5
		0.99: 0.001, // 0.98.9-0.99.1,
	}
)

func Milliseconds(duration time.Duration) float64 {
	value := duration.Milliseconds()
	if value == 0 {
		return 1
	}
	return float64(value)
}
