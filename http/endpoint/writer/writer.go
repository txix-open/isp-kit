package writer

import (
	"bytes"
	"net/http"
	"sync"
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

var pool = sync.Pool{
	New: func() interface{} {
		return NewMiddlewareWriter()
	},
}

func Acquire(upstream http.ResponseWriter) *MiddlewareWriter {
	var writer *MiddlewareWriter
	var ok bool
	writer, ok = pool.Get().(*MiddlewareWriter)
	if !ok {
		writer = NewMiddlewareWriter()
	}
	writer.PrepareToWork(upstream)
	return writer
}

func Release(w *MiddlewareWriter) {
	pool.Put(w)
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
