package writer

import (
	"bytes"
	"net/http"
)

type MiddlewareWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func NewMiddlewareWriter() *MiddlewareWriter {
	return &MiddlewareWriter{body: bytes.NewBuffer(make([]byte, 1024))}
}

func (m *MiddlewareWriter) PrepareToWork(upstream http.ResponseWriter) {
	m.ResponseWriter = upstream
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

func (m *MiddlewareWriter) GetBody() []byte {
	return m.body.Bytes()
}
