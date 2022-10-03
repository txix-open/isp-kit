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
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "client_request_duration_ms",
			Help:       "The latencies of calling external services via GRPC",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.SummaryVec)
	return s
}

func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}
