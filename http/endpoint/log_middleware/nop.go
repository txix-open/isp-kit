package log_middleware

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
)

func Noop() endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return next(ctx, w, r)
		}
	}
}
