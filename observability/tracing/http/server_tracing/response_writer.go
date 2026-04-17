package server_tracing

import (
	"bufio"
	"net"
	"net/http"

	"github.com/pkg/errors"
)

// scSource is an interface for HTTP response writers that provide status code access.
type scSource interface {
	http.ResponseWriter
	// StatusCode returns the HTTP status code.
	StatusCode() int
}

// writerWrapper wraps an http.ResponseWriter to capture the status code.
type writerWrapper struct {
	http.ResponseWriter

	// statusCode holds the captured HTTP status code.
	statusCode int
}

// Hijack implements the http.Hijacker interface by delegating to the upstream response writer.
// It returns an error if the upstream does not support hijacking.
func (w *writerWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	upstream, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("writerWrapper: upstream writer doesn't implement Hijack")
	}
	return upstream.Hijack()
}

// StatusCode returns the captured HTTP status code, or http.StatusOK if not yet set.
func (w *writerWrapper) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

// WriteHeader captures the status code and delegates to the upstream response writer.
func (w *writerWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
