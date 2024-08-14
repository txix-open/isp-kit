package http_metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
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
	duration          *prometheus.SummaryVec
	dnsLookup         *prometheus.SummaryVec
	connEstablishment *prometheus.SummaryVec
	requestReading    *prometheus.SummaryVec
	responseWriting   *prometheus.SummaryVec
}

func NewClientStorage(reg *metrics.Registry) *ClientStorage {
	s := &ClientStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_request_duration",
			Help:       "The latencies of calling external services via HTTP",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		connEstablishment: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "http_client_connect_duration",
			Help:       "The latencies of connection establishment",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		dnsLookup: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "http_client_dns_duration",
			Help:       "The latencies of DNS lookup",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		responseWriting: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "http_client_response_write_duration",
			Help:       "The latencies of response writing",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),
	}
	return s
}

func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}

func (s *ClientStorage) ObserveConnEstablishment(endpoint string, duration time.Duration) {
	s.connEstablishment.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}

func (s *ClientStorage) ObserveDnsLookup(endpoint string, duration time.Duration) {
	s.dnsLookup.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}

func (s *ClientStorage) ObserveResponseWriting(endpoint string, duration time.Duration) {
	s.responseWriting.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}
