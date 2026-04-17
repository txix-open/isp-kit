package http_metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type serverEndpointContextKey struct{}

// nolint:gochecknoglobals
var (
	serverEndpointContextKeyValue = serverEndpointContextKey{}
)

// ServerEndpointToContext stores the HTTP endpoint name in the context.
// This is useful for passing endpoint information through middleware chains.
func ServerEndpointToContext(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, serverEndpointContextKeyValue, endpoint)
}

// ServerEndpoint retrieves the HTTP endpoint name from the context.
// Returns an empty string if no endpoint was set.
func ServerEndpoint(ctx context.Context) string {
	s, _ := ctx.Value(serverEndpointContextKeyValue).(string)
	return s
}

// ServerStorage collects metrics for HTTP server operations, including request
// latency, request/response body sizes, and HTTP status code counts.
type ServerStorage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

// NewServerStorage creates a new ServerStorage instance and registers its metrics
// with the provided registry. The metrics are labeled by HTTP method and endpoint.
func NewServerStorage(reg *metrics.Registry) *ServerStorage {
	s := &ServerStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_duration_ms",
			Help:       "The latency of the HTTP requests",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "endpoint"})),
		requestBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "request_body_size",
			Help:       "The size of request body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "endpoint"})),
		responseBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "response_body_size",
			Help:       "The size of response body",
			Objectives: metrics.DefaultObjectives,
		}, []string{"method", "endpoint"})),
		statusCounter: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem:   "http",
			Name:        "status_code_count",
			Help:        "Counter of statuses codes",
			ConstLabels: nil,
		}, []string{"method", "endpoint", "code"})),
	}
	return s
}

// ObserveDuration records the latency of an HTTP request, labeled by method and endpoint.
func (s *ServerStorage) ObserveDuration(method string, endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(method, endpoint).Observe(metrics.Milliseconds(duration))
}

// ObserveRequestBodySize records the size of an HTTP request body in bytes.
func (s *ServerStorage) ObserveRequestBodySize(method string, endpoint string, size int) {
	s.requestBodySize.WithLabelValues(method, endpoint).Observe(float64(size))
}

// ObserveResponseBodySize records the size of an HTTP response body in bytes.
func (s *ServerStorage) ObserveResponseBodySize(method string, endpoint string, size int) {
	s.responseBodySize.WithLabelValues(method, endpoint).Observe(float64(size))
}

// CountStatusCode increments the counter for a specific HTTP status code,
// labeled by method, endpoint, and status code.
func (s *ServerStorage) CountStatusCode(method string, endpoint string, code int) {
	s.statusCounter.WithLabelValues(method, endpoint, strconv.Itoa(code)).Inc()
}
