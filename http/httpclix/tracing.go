package httpclix

import (
	metrics "github.com/txix-open/isp-kit/metrics/http_metrics"
	"net/http/httptrace"
	"time"
)

type ClientTracer struct {
	clientStorage              *metrics.ClientStorage
	connEstablishmentStartTime time.Time
	dnsStartTime               time.Time
	requestWritingStartTime    time.Time
	responseReadingStartTime   time.Time
	endpoint                   string
}

func NewClientTracer(clientStorage *metrics.ClientStorage, endpoint string) *ClientTracer {
	return &ClientTracer{
		clientStorage: clientStorage,
		endpoint:      endpoint,
	}
}

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

func (cli *ClientTracer) ResponseReceived() {
	if cli.responseReadingStartTime.IsZero() {
		return
	}

	respReadingDur := time.Since(cli.responseReadingStartTime)
	cli.clientStorage.ObserveResponseReading(cli.endpoint, respReadingDur)
}
