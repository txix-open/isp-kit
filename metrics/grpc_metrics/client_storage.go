package grpc_metrics

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
			Subsystem:  "grpc",
			Name:       "client_request_duration_ms",
			Help:       "The latencies of calling external services via GRPC",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
	}
	return s
}

func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}
