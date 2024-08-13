package httpclix

import (
	metrics "github.com/txix-open/isp-kit/metrics/http_metrics"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"time"
)

type ClientTracer struct {
	clientStorage *metrics.ClientStorage
	connEstablishmentStart,
	dnsStart,
	requestReadingStart,
	responseWritingStart time.Time
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
			cli.clientStorage.ObserveDnsLookup(cli.endpoint, cli.dnsHost, dnsLookupDur)
		},

		// taking into account conn pooling + dialing
		GetConn: func(hostPort string) {
			cli.connEstablishmentStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connEstablishmentDur := time.Since(cli.connEstablishmentStart)
			cli.clientStorage.ObserveConnEstablishment(cli.endpoint, network, addr, connEstablishmentDur)
		},

		// client stars writing the body, hence server starts reading it
		Got100Continue: func() {
			cli.requestReadingStart = time.Now()
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			if code == http.StatusProcessing {
				// done reading the request body
				requestReadingDur := time.Since(cli.requestReadingStart)
				cli.clientStorage.ObserveRequestReading(cli.endpoint, requestReadingDur)
			}
			return nil
		},

		// response writing started
		GotFirstResponseByte: func() {
			cli.responseWritingStart = time.Now()
		},
	}
	return &tracingCli
}

func (cli *ClientTracer) ResponseReceived() {
	if cli.responseWritingStart.IsZero() {
		return
	}

	respWritingDur := time.Since(cli.responseWritingStart)
	cli.clientStorage.ObserveResponseWriting(cli.endpoint, respWritingDur)
}
