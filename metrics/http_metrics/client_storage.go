package http_metrics

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type ClientStorage struct {
	duration *prometheus.SummaryVec
}

func NewClientStorage(reg *metrics.Registry) *ClientStorage {
	s := &ClientStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_request_duration_ms",
			Help:       "The latencies of calling external services via HTTP",
			Objectives: metrics.DefaultObjectives,
		}, []string{"url"})),
	}
	return s
}

func (s *ClientStorage) ObserveDuration(url string, duration time.Duration) {
	s.duration.WithLabelValues(url).Observe(float64(duration.Milliseconds()))
}
