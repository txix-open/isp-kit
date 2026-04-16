package httpcli

import (
	"net/http"
	"time"
)

// Request wraps an http.Request with additional metadata for middleware processing.
type Request struct {
	Raw *http.Request

	retryOptions *retryOptions
	body         []byte
	timeout      time.Duration
}

// Body returns the request body as bytes.
//
// Returns an empty slice if MultipartRequestBody was used, as multipart data
// is streamed directly and not buffered.
func (r *Request) Body() []byte {
	return r.body
}
