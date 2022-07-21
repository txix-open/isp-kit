package http_metrics

import (
	"strconv"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Storage struct {
	duration         *prometheus.HistogramVec
	requestBodySize  *prometheus.HistogramVec
	responseBodySize *prometheus.HistogramVec
}

func NewStorage(reg *metrics.Registry) *Storage {
	s := &Storage{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_duration_ms",
			Help:      "The latency of the HTTP requests",
			Buckets:   metrics.DefaultDurationMsBuckets,
		}, []string{"method", "path", "code"}),
		requestBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_body_size",
			Help:      "The size of request body",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"method", "path"}),
		responseBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "response_body_size",
			Help:      "The size of response body",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"method", "path"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.HistogramVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.HistogramVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.HistogramVec)
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
