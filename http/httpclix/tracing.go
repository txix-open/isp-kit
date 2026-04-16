package httpclix

import (
	metrics "github.com/txix-open/isp-kit/metrics/http_metrics"
	"net/http/httptrace"
	"time"
)

// ClientTracer collects detailed timing information for HTTP client operations.
//
// It implements the httptrace.ClientTrace interface to measure DNS lookup,
// connection establishment, request writing, and response reading durations.
type ClientTracer struct {
	clientStorage              *metrics.ClientStorage
	connEstablishmentStartTime time.Time
	dnsStartTime               time.Time
	requestWritingStartTime    time.Time
	responseReadingStartTime   time.Time
	endpoint                   string
}

// NewClientTracer creates a new ClientTracer for the given endpoint.
func NewClientTracer(clientStorage *metrics.ClientStorage, endpoint string) *ClientTracer {
	return &ClientTracer{
		clientStorage: clientStorage,
		endpoint:      endpoint,
	}
}

// ClientTrace returns an httptrace.ClientTrace that records timing metrics
// for the various phases of an HTTP request.
func (cli *ClientTracer) ClientTrace() *httptrace.ClientTrace {
	tracingCli := httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			cli.dnsStartTime = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if cli.dnsStartTime.IsZero() {
				return
			}
			dnsLookupDur := time.Since(cli.dnsStartTime)
			cli.clientStorage.ObserveDnsLookup(cli.endpoint, dnsLookupDur)
		},

		ConnectStart: func(network string, addr string) {
			cli.connEstablishmentStartTime = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if cli.connEstablishmentStartTime.IsZero() {
				return
			}
			connEstablishmentDur := time.Since(cli.connEstablishmentStartTime)
			cli.clientStorage.ObserveConnEstablishment(cli.endpoint, connEstablishmentDur)
		},

		// starting to write body
		WroteHeaders: func() {
			cli.requestWritingStartTime = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if cli.requestWritingStartTime.IsZero() {
				return
			}
			requestWritingDur := time.Since(cli.requestWritingStartTime)
			cli.clientStorage.ObserveRequestWriting(cli.endpoint, requestWritingDur)
		},

		// response writing started
		GotFirstResponseByte: func() {
			cli.responseReadingStartTime = time.Now()
		},
	}
	return &tracingCli
}

// ResponseReceived is called when the response body has been fully read.
//
// Records the duration of response reading.
func (cli *ClientTracer) ResponseReceived() {
	if cli.responseReadingStartTime.IsZero() {
		return
	}

	respReadingDur := time.Since(cli.responseReadingStartTime)
	cli.clientStorage.ObserveResponseReading(cli.endpoint, respReadingDur)
}
