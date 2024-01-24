package server_tracing

import (
	"bufio"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

type scSource interface {
	http.ResponseWriter
	StatusCode() int
}

type writerWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *writerWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	upstream, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("writerWrapper: upstream writer doesn't implement Hijack")
	}
	return upstream.Hijack()
}

func (w *writerWrapper) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *writerWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
