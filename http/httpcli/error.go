package httpcli

import (
	"fmt"
	"net/url"
)

// nolint:errname
type ErrorResponse struct {
	Url        *url.URL
	StatusCode int
	Body       []byte
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("http call error: url=%s status_code=%d, body=%s", e.Url.String(), e.StatusCode, e.Body)
}
