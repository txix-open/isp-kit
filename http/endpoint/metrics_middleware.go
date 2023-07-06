package endpoint

import (
	"context"
	"net/http"
	"time"

	"github.com/integration-system/isp-kit/http/endpoint/buffer"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
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

func Metrics(storage *http_metrics.ServerStorage) Middleware {
	return func(next HandlerFunc) HandlerFunc {
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
