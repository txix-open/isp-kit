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
}

func NewStorage(reg *metrics.Registry) *Storage {
	s := &Storage{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "grpc",
			Name:       "request_duration_ms",
			Help:       "The latency of the GRPC requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint", "code"}),
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
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.SummaryVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.SummaryVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.SummaryVec)
	return s
}

func (s *Storage) ObserveDuration(endpoint string, statusCode codes.Code, duration time.Duration) {
	s.duration.WithLabelValues(endpoint, statusCode.String()).Observe(float64(duration.Milliseconds()))
}

func (s *Storage) ObserveRequestBodySize(endpoint string, size int) {
	s.requestBodySize.WithLabelValues(endpoint).Observe(float64(size))
}

func (s *Storage) ObserveResponseBodySize(endpoint string, size int) {
	s.responseBodySize.WithLabelValues(endpoint).Observe(float64(size))
}
