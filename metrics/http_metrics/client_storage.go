package http_metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.txix.ru/isp/isp-kit/metrics"
)

type clientEndpointContextKey struct{}

var (
	clientEndpointContextKeyValue = clientEndpointContextKey{}
)

func ClientEndpointToContext(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, clientEndpointContextKeyValue, endpoint)
}

func ClientEndpoint(ctx context.Context) string {
	s, _ := ctx.Value(clientEndpointContextKeyValue).(string)
	return s
}

type ClientStorage struct {
	duration *prometheus.SummaryVec
}

func NewClientStorage(reg *metrics.Registry) *ClientStorage {
	s := &ClientStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_request_duration_ms",
			Help:       "The latencies of calling external services via HTTP",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
	}
	return s
}

func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}
