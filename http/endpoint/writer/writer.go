package writer

import (
	"bytes"
	"net/http"
)

type MiddlewareWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func NewMiddlewareWriter() *MiddlewareWriter {
	return &MiddlewareWriter{body: bytes.NewBuffer(make([]byte, 1024))}
}

func (m *MiddlewareWriter) PrepareToWork(upstream http.ResponseWriter) {
	m.ResponseWriter = upstream
	m.statusCode = 0
	m.body.Reset()
}

func (m *MiddlewareWriter) Write(b []byte) (int, error) {
	n, err := m.ResponseWriter.Write(b)
	if err != nil {
		return n, err
	}

	n, err = m.body.Write(b)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (m *MiddlewareWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
	m.ResponseWriter.WriteHeader(statusCode)
}

func (m *MiddlewareWriter) GetBody() []byte {
	return m.body.Bytes()
}

func (m *MiddlewareWriter) StatusCode() int {
	if m.statusCode == 0 {
		return http.StatusOK
	}
	return m.statusCode
}
