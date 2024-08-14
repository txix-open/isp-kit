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
	requestWriting    *prometheus.SummaryVec
	responseReading   *prometheus.SummaryVec
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
			Name:       "client_connect_duration",
			Help:       "The latencies of connection establishment",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		requestWriting: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_request_write_duration",
			Help:       "The latencies of request writing",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		dnsLookup: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_dns_duration",
			Help:       "The latencies of DNS lookup",
			Objectives: metrics.DefaultObjectives,
		}, []string{"endpoint"})),

		responseReading: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_response_read_duration",
			Help:       "The latencies of response reading",
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

func (s *ClientStorage) ObserveRequestWriting(endpoint string, duration time.Duration) {
	s.requestWriting.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}

func (s *ClientStorage) ObserveDnsLookup(endpoint string, duration time.Duration) {
	s.dnsLookup.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}

func (s *ClientStorage) ObserveResponseReading(endpoint string, duration time.Duration) {
	s.responseReading.WithLabelValues(endpoint).Observe(float64(duration.Nanoseconds()))
}
