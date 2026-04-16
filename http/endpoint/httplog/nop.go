package httplog

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
)

// Noop returns a no-operation logging middleware that simply passes through requests.
// It is useful as a placeholder or for testing when logging is not needed.
func Noop() endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return next(ctx, w, r)
		}
	}
}
