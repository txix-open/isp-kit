package endpoint

import (
	"context"
	"net/http"
	"time"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint/buffer"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
)

type scSource interface {
	StatusCode() int
}

type writerWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *writerWrapper) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *writerWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Metrics(storage *http_metrics.ServerStorage) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			endpoint := http_metrics.ServerEndpoint(r.Context())
			if endpoint == "" {
				return next(ctx, w, r)
			}

			var scSrc scSource
			buf, isBuffer := w.(*buffer.Buffer)
			if isBuffer {
				scSrc = buf
			} else {
				scSrc = &writerWrapper{ResponseWriter: w}
			}

			start := time.Now()
			err := next(ctx, w, r)
			storage.ObserveDuration(r.Method, endpoint, time.Since(start))
			storage.CountStatusCode(r.Method, endpoint, scSrc.StatusCode())
			if isBuffer {
				storage.ObserveRequestBodySize(r.Method, endpoint, len(buf.RequestBody()))
				storage.ObserveResponseBodySize(r.Method, endpoint, len(buf.ResponseBody()))
			}

			return err
		}
	}
}
