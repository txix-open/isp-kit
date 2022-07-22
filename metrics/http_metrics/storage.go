package http_metrics

import (
	"strconv"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Storage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
}

func NewStorage(reg *metrics.Registry) *Storage {
	s := &Storage{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_duration_ms",
			Help:       "The latency of the HTTP requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "path", "code"}),
		requestBodySize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_body_size",
			Help:       "The size of request body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "path"}),
		responseBodySize: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "response_body_size",
			Help:       "The size of response body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "path"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.SummaryVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.SummaryVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.SummaryVec)
	return s
}

func (s *Storage) ObserveDuration(method string, path string, statusCode int, duration time.Duration) {
	s.duration.WithLabelValues(method, path, strconv.Itoa(statusCode)).Observe(float64(duration.Milliseconds()))
}

func (s *Storage) ObserveRequestBodySize(method string, path string, size int) {
	s.requestBodySize.WithLabelValues(method, path).Observe(float64(size))
}

func (s *Storage) ObserveResponseBodySize(method string, path string, size int) {
	s.responseBodySize.WithLabelValues(method, path).Observe(float64(size))
}
