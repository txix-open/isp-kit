package buffer

import (
	"net/http"
	"sync"
)

// pool is a global sync.Pool for reusing Buffer instances.
// It reduces memory allocation overhead by pooling Buffer objects.
var pool = sync.Pool{
	New: func() any {
		return New()
	},
}

// Acquire retrieves a Buffer from the pool and initializes it with the given ResponseWriter.
// It is safe for concurrent use and should be paired with Release.
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

// Release returns a Buffer to the pool for reuse.
// It should be called after the Buffer is no longer needed.
func Release(w *Buffer) {
	pool.Put(w)
}
