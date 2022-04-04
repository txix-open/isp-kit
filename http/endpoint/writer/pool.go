package writer

import (
	"net/http"
	"sync"
)

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
