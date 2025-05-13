package httpcli

import (
	"bytes"
	"sync"
)

// nolint:gochecknoglobals,mnd
var (
	bpool = sync.Pool{New: func() any {
		return bytes.NewBuffer(make([]byte, 1024))
	}}
)

func acquireBuffer() *bytes.Buffer {
	buf := bpool.Get().(*bytes.Buffer) // nolint:forcetypeassert
	buf.Reset()
	return buf
}

func releaseBuffer(buf *bytes.Buffer) {
	bpool.Put(buf)
}
