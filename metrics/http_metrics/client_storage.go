package http_metrics

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type clientEndpointContextKey struct{}

// nolint:lll,gochecknoglobals
var (
	clientEndpointContextKeyValue = clientEndpointContextKey{}

	ipPortRegexp   = regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?::\d+)?\b`)
	ipv6PortRegexp = regexp.MustCompile(`\[([0-9a-fA-F:]+)\]:(\d+)`)
	urlPattern     = regexp.MustCompile(`(https?://\S+|ftp://\S+|file://\S+|[a-zA-Z0-9_-]+\.([a-zA-Z]{2,})+\S+)`)
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

	statusCounter *prometheus.CounterVec
	errorCounter  *prometheus.CounterVec
}

func NewClientStorage(reg *metrics.Registry) *ClientStorage {
	s := &ClientStorage{
		duration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "http",
			Name:       "client_request_duration_ms",
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

		statusCounter: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "client_status_code_count",
			Help:      "Counter of HTTP status codes",
		}, []string{"endpoint", "code"})),

		errorCounter: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "client_error_count",
			Help:      "Counter of HTTP client errors",
		}, []string{"endpoint", "error"})),
	}
	return s
}

func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}

func (s *ClientStorage) ObserveConnEstablishment(endpoint string, duration time.Duration) {
	s.connEstablishment.WithLabelValues(endpoint).Observe(float64(duration))
}

func (s *ClientStorage) ObserveRequestWriting(endpoint string, duration time.Duration) {
	s.requestWriting.WithLabelValues(endpoint).Observe(float64(duration))
}

func (s *ClientStorage) ObserveDnsLookup(endpoint string, duration time.Duration) {
	s.dnsLookup.WithLabelValues(endpoint).Observe(float64(duration))
}

func (s *ClientStorage) ObserveResponseReading(endpoint string, duration time.Duration) {
	s.responseReading.WithLabelValues(endpoint).Observe(float64(duration))
}

func (s *ClientStorage) CountStatusCode(endpoint string, code int) {
	s.statusCounter.WithLabelValues(endpoint, strconv.Itoa(code)).Inc()
}

func (s *ClientStorage) CountError(endpoint string, err error) {
	s.errorCounter.WithLabelValues(endpoint, trimError(err)).Inc()
}

func trimError(err error) string {
	text := err.Error()
	text = urlPattern.ReplaceAllString(text, "%URL%")
	text = ipPortRegexp.ReplaceAllString(text, "xxx.xxx.xxx.xxx:xxxx")
	text = ipv6PortRegexp.ReplaceAllString(text, "[xxx:xxx:xxx:xxx]:xxxx")

	return text
}
