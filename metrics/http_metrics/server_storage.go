package http_metrics

import (
	"strconv"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type ServerStorage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

func NewServerStorage(reg *metrics.Registry) *ServerStorage {
	s := &ServerStorage{
		duration: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_duration_ms",
			Help:       "The latency of the HTTP requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "path"}),
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
		statusCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem:   "http",
			Name:        "status_code_count",
			Help:        "Counter of statuses codes",
			ConstLabels: nil,
		}, []string{"method", "path", "code"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.SummaryVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.SummaryVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.SummaryVec)
	s.statusCounter = reg.GetOrRegister(s.statusCounter).(*prometheus.CounterVec)
	return s
}

func (s *ServerStorage) ObserveDuration(method string, path string, duration time.Duration) {
	s.duration.WithLabelValues(method, path).Observe(float64(duration.Milliseconds()))
}

func (s *ServerStorage) ObserveRequestBodySize(method string, path string, size int) {
	s.requestBodySize.WithLabelValues(method, path).Observe(float64(size))
}

func (s *ServerStorage) ObserveResponseBodySize(method string, path string, size int) {
	s.responseBodySize.WithLabelValues(method, path).Observe(float64(size))
}

func (s *ServerStorage) CountStatusCode(method string, path string, code int) {
	s.statusCounter.WithLabelValues(method, path, strconv.Itoa(code)).Inc()
}
