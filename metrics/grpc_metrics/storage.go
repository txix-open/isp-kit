package grpc_metrics

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

type Storage struct {
	duration         *prometheus.HistogramVec
	requestBodySize  *prometheus.HistogramVec
	responseBodySize *prometheus.HistogramVec
}

func NewStorage(reg *metrics.Registry) *Storage {
	s := &Storage{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "grpc",
			Name:      "request_duration_ms",
			Help:      "The latency of the GRPC requests",
			Buckets:   metrics.DefaultDurationMsBuckets,
		}, []string{"endpoint", "code"}),
		requestBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "grpc",
			Name:      "request_body_size",
			Help:      "The size of request body",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"endpoint"}),
		responseBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "grpc",
			Name:      "response_body_size",
			Help:      "The size of response body",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"endpoint"}),
	}
	s.duration = reg.GetOrRegister(s.duration).(*prometheus.HistogramVec)
	s.requestBodySize = reg.GetOrRegister(s.requestBodySize).(*prometheus.HistogramVec)
	s.responseBodySize = reg.GetOrRegister(s.responseBodySize).(*prometheus.HistogramVec)
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
