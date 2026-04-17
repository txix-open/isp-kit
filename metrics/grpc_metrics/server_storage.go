package grpc_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
	"google.golang.org/grpc/codes"
)

// ServerStorage collects metrics for gRPC server operations, including request
// latency, request/response body sizes, and gRPC status code counts.
type ServerStorage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

// NewServerStorage creates a new ServerStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by gRPC endpoint.
func NewServerStorage(reg *metrics.Registry) *ServerStorage {
	s := &ServerStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "request_duration_ms",
			Help:       "The latency of the GRPC requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
		requestBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "request_body_size",
			Help:       "The size of request body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
		responseBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "response_body_size",
			Help:       "The size of response body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
		statusCounter: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "status_code_count",
			Help:      "Counter of statuses codes",
		}, []string{"endpoint", "code"})),
	}
	return s
}

// ObserveDuration records the latency of a gRPC request, labeled by endpoint.
func (s *ServerStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}

// ObserveRequestBodySize records the size of a gRPC request payload in bytes.
func (s *ServerStorage) ObserveRequestBodySize(endpoint string, size int) {
	s.requestBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

// ObserveResponseBodySize records the size of a gRPC response payload in bytes.
func (s *ServerStorage) ObserveResponseBodySize(endpoint string, size int) {
	s.responseBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

// CountStatusCode increments the counter for a specific gRPC status code,
// labeled by endpoint and status code.
func (s *ServerStorage) CountStatusCode(endpoint string, code codes.Code) {
	s.statusCounter.WithLabelValues(endpoint, code.String()).Inc()
}
