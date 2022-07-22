package grpc_metrics

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type Storage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

func NewStorage(reg *metrics.Registry) *Storage {
	s := &Storage{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "request_duration_ms",
			Help:       "The latency of the GRPC requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"}),
		requestBodySize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "request_body_size",
			Help:       "The size of request body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"}),
		responseBodySize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "response_body_size",
			Help:       "The size of response body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"}),
		statusCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "grpc",
			Name:      "status_code_count",
			Help:      "Counter of statuses codes",
		}, []string{"endpoint", "code"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.SummaryVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.SummaryVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.SummaryVec)
	s.statusCounter = reg.GetOrRegister(s.statusCounter).(*prometheus.CounterVec)
	return s
}

func (s *Storage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(float64(duration.Milliseconds()))
}

func (s *Storage) ObserveRequestBodySize(endpoint string, size int) {
	s.requestBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

func (s *Storage) ObserveResponseBodySize(endpoint string, size int) {
	s.responseBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

func (s *Storage) CountStatusCode(endpoint string, code codes.Code) {
	s.statusCounter.WithLabelValues(endpoint, code.String()).Inc()
}
