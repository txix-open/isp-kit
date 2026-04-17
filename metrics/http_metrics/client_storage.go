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

	ipPortRegexp   = regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	ipv6PortRegexp = regexp.MustCompile(`\[([0-9a-fA-F:]+)\]:(\d+)`)
	urlPattern     = regexp.MustCompile(`(https?://\S+|ftp://\S+|file://\S+|[a-zA-Z0-9_-]+\.([a-zA-Z]{2,})+\S+)`)
)

// ClientEndpointToContext stores the HTTP client endpoint name in the context.
func ClientEndpointToContext(ctx context.Context, endpoint string) context.Context {
	return context.WithValue(ctx, clientEndpointContextKeyValue, endpoint)
}

// ClientEndpoint retrieves the HTTP client endpoint name from the context.
// Returns an empty string if no endpoint was set.
func ClientEndpoint(ctx context.Context) string {
	s, _ := ctx.Value(clientEndpointContextKeyValue).(string)
	return s
}

// ClientStorage collects metrics for HTTP client operations, including request
// latency, connection establishment, DNS lookup, and error counts.
type ClientStorage struct {
	duration          *prometheus.SummaryVec
	dnsLookup         *prometheus.SummaryVec
	connEstablishment *prometheus.SummaryVec
	requestWriting    *prometheus.SummaryVec
	responseReading   *prometheus.SummaryVec

	statusCounter *prometheus.CounterVec
	errorCounter  *prometheus.CounterVec
}

// NewClientStorage creates a new ClientStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by the client endpoint.
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

// ObserveDuration records the total duration of an HTTP client request.
func (s *ClientStorage) ObserveDuration(endpoint string, duration time.Duration) {
	s.duration.WithLabelValues(endpoint).Observe(metrics.Milliseconds(duration))
}

// ObserveConnEstablishment records the time taken to establish a connection.
func (s *ClientStorage) ObserveConnEstablishment(endpoint string, duration time.Duration) {
	s.connEstablishment.WithLabelValues(endpoint).Observe(float64(duration))
}

// ObserveRequestWriting records the time taken to write the request body.
func (s *ClientStorage) ObserveRequestWriting(endpoint string, duration time.Duration) {
	s.requestWriting.WithLabelValues(endpoint).Observe(float64(duration))
}

// ObserveDnsLookup records the time taken for DNS resolution.
func (s *ClientStorage) ObserveDnsLookup(endpoint string, duration time.Duration) {
	s.dnsLookup.WithLabelValues(endpoint).Observe(float64(duration))
}

// ObserveResponseReading records the time taken to read the response.
func (s *ClientStorage) ObserveResponseReading(endpoint string, duration time.Duration) {
	s.responseReading.WithLabelValues(endpoint).Observe(float64(duration))
}

// CountStatusCode increments the counter for a specific HTTP status code.
func (s *ClientStorage) CountStatusCode(endpoint string, code int) {
	s.statusCounter.WithLabelValues(endpoint, strconv.Itoa(code)).Inc()
}

// CountError increments the error counter for a specific error type.
// The error message is sanitized to remove sensitive information like URLs and IPs.
func (s *ClientStorage) CountError(endpoint string, err error) {
	s.errorCounter.WithLabelValues(endpoint, trimError(err)).Inc()
}

// trimError sanitizes an error message by replacing URLs, IP addresses, and
// IPv6 addresses with placeholders to avoid exposing sensitive information in metrics.
func trimError(err error) string {
	text := err.Error()
	text = urlPattern.ReplaceAllString(text, "%URL%")
	text = ipPortRegexp.ReplaceAllString(text, "xxx.xxx.xxx.xxx:xxxx")
	text = ipv6PortRegexp.ReplaceAllString(text, "[xxx:xxx:xxx:xxx]:xxxx")

	return text
}
