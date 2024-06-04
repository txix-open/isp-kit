package grpc_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"google.golang.org/grpc/codes"
)

type ServerStorage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

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

func (s *ServerStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}

func (s *ServerStorage) ObserveRequestBodySize(endpoint string, size int) {
	s.requestBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

func (s *ServerStorage) ObserveResponseBodySize(endpoint string, size int) {
	s.responseBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

func (s *ServerStorage) CountStatusCode(endpoint string, code codes.Code) {
	s.statusCounter.WithLabelValues(endpoint, code.String()).Inc()
}
