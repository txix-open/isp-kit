package http_metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type serverEndpointContextKey struct{}

var (
	serverEndpointContextKeyValue = serverEndpointContextKey{}
)

func ServerEndpointToContext(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, serverEndpointContextKeyValue, endpoint)
}

func ServerEndpoint(ctx context.Context) string {
	s, _ := ctx.Value(serverEndpointContextKeyValue).(string)
	return s
}

type ServerStorage struct {
	duration         *prometheus.SummaryVec
	requestBodySize  *prometheus.SummaryVec
	responseBodySize *prometheus.SummaryVec
	statusCounter    *prometheus.CounterVec
}

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

func (s *ServerStorage) ObserveDuration(method string, endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(method, endpoint).Observe(float64(duration.Milliseconds()))
}

func (s *ServerStorage) ObserveRequestBodySize(method string, endpoint string, size int) {
	s.requestBodySize.WithLabelValues(method, endpoint).Observe(float64(size))
}

func (s *ServerStorage) ObserveResponseBodySize(method string, endpoint string, size int) {
	s.responseBodySize.WithLabelValues(method, endpoint).Observe(float64(size))
}

func (s *ServerStorage) CountStatusCode(method string, endpoint string, code int) {
	s.statusCounter.WithLabelValues(method, endpoint, strconv.Itoa(code)).Inc()
}
