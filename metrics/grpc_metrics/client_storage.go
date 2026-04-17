package grpc_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

// ClientStorage collects metrics for gRPC client operations, primarily focusing
// on request latency when calling external gRPC services.
type ClientStorage struct {
	duration *prometheus.SummaryVec
}

// NewClientStorage creates a new ClientStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by the gRPC endpoint.
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

// ObserveDuration records the latency of a gRPC client request, labeled by endpoint.
func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}
