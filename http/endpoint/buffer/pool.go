package buffer

import (
	"net/http"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return New()
	},
}

func Acquire(w http.ResponseWriter) *Buffer {
	var buffer *Buffer
	var ok bool
	buffer, ok = pool.Get().(*Buffer)
	if !ok {
		buffer = New()
	}
	buffer.Reset(w)
	return buffer
}

func Release(w *Buffer) {
	pool.Put(w)
}
