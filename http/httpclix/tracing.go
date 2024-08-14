package httpclix

import (
	metrics "github.com/txix-open/isp-kit/metrics/http_metrics"
	"net/http/httptrace"
	"time"
)

type ClientTracer struct {
	clientStorage *metrics.ClientStorage
	connEstablishmentStart,
	dnsStart,
	requestWritingStart,
	responseReadingStart time.Time
	dnsHost,
	endpoint string
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
			cli.dnsStart = time.Now()
			cli.dnsHost = info.Host
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsLookupDur := time.Since(cli.dnsStart)
			cli.clientStorage.ObserveDnsLookup(cli.endpoint, dnsLookupDur)
		},

		// taking into account conn pooling + dialing
		GetConn: func(hostPort string) {
			cli.connEstablishmentStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connEstablishmentDur := time.Since(cli.connEstablishmentStart)
			cli.clientStorage.ObserveConnEstablishment(cli.endpoint, connEstablishmentDur)
		},

		// starting to write body
		WroteHeaders: func() {
			cli.requestWritingStart = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			requestWritingDur := time.Since(cli.requestWritingStart)
			cli.clientStorage.ObserveRequestWriting(cli.endpoint, requestWritingDur)
		},

		// response writing started
		GotFirstResponseByte: func() {
			cli.responseReadingStart = time.Now()
		},
	}
	return &tracingCli
}

func (cli *ClientTracer) ResponseReceived() {
	if cli.responseReadingStart.IsZero() {
		return
	}

	respReadingDur := time.Since(cli.responseReadingStart)
	cli.clientStorage.ObserveResponseReading(cli.endpoint, respReadingDur)
}
