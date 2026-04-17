package buffer

import (
	"bytes"
)

// RequestBody wraps a bytes.Buffer to provide an io.ReadCloser for request body replay.
// It is used when the request body needs to be read multiple times (e.g., for logging).
type RequestBody struct {
	body *bytes.Buffer
}

// NewRequestBody creates a new RequestBody from a byte slice.
func NewRequestBody(body []byte) RequestBody {
	return RequestBody{
		body: bytes.NewBuffer(body),
	}
}

// Read reads data from the underlying buffer into p.
// It returns the number of bytes read and any error encountered.
func (r RequestBody) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

// Close is a no-op to satisfy the io.ReadCloser interface.
func (r RequestBody) Close() error {
	return nil
}
