// Package buffer provides a ResponseWriter wrapper that buffers request and response bodies.
// It is used for logging and metrics collection in HTTP middleware.
package buffer

import (
	"bytes"
	"io"
	"net/http"
)

const (
	// defaultBufferSize is the initial buffer size in bytes.
	defaultBufferSize = 1024
)

// Buffer wraps http.ResponseWriter to capture request and response bodies.
// It is commonly used in middleware for logging and metrics collection.
type Buffer struct {
	http.ResponseWriter

	requestBuffer  *bytes.Buffer
	responseBuffer *bytes.Buffer
	statusCode     int
}

// New creates a new Buffer with initialized request and response buffers.
func New() *Buffer {
	return &Buffer{
		requestBuffer:  bytes.NewBuffer(make([]byte, defaultBufferSize)),
		responseBuffer: bytes.NewBuffer(make([]byte, defaultBufferSize)),
	}
}

// Reset reinitializes the buffer with a new ResponseWriter and clears internal state.
// It should be called before reusing a pooled Buffer.
func (m *Buffer) Reset(w http.ResponseWriter) {
	m.ResponseWriter = w
	m.statusCode = 0
	m.responseBuffer.Reset()
	m.requestBuffer.Reset()
}

// Write writes data to both the underlying ResponseWriter and the response buffer.
// It returns the number of bytes written and any error encountered.
func (m *Buffer) Write(b []byte) (int, error) {
	n, err := m.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}

	n, err = m.responseBuffer.Write(b)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// WriteHeader captures the status code and delegates to the underlying ResponseWriter.
func (m *Buffer) WriteHeader(statusCode int) {
	m.statusCode = statusCode
	m.ResponseWriter.WriteHeader(statusCode)
}

// ResponseBody returns the captured response body as a byte slice.
func (m *Buffer) ResponseBody() []byte {
	return m.responseBuffer.Bytes()
}

// RequestBody returns the captured request body as a byte slice.
func (m *Buffer) RequestBody() []byte {
	return m.requestBuffer.Bytes()
}

// ReadRequestBody reads the entire request body into the buffer.
// It returns an error if the read operation fails.
func (m *Buffer) ReadRequestBody(r io.Reader) error {
	_, err := io.Copy(m.requestBuffer, r)
	return err
}

// StatusCode returns the captured status code, or http.StatusOK if not set.
func (m *Buffer) StatusCode() int {
	if m.statusCode == 0 {
		return http.StatusOK
	}
	return m.statusCode
}
