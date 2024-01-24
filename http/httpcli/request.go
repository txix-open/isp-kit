package httpcli

import (
	"net/http"
	"time"
)

type Request struct {
	Raw *http.Request

	retryOptions *retryOptions
	body         []byte
	timeout      time.Duration
}

// Body
// Returns request body in bytes
// Always returns empty slice if you use RequestBuilder.MultipartRequestBody
func (r *Request) Body() []byte {
	return r.body
}
