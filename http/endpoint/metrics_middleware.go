package endpoint

import (
	"context"
	"net/http"
	"time"

	"github.com/integration-system/isp-kit/http/endpoint/buffer"
)

type MetricStorage interface {
	ObserveDuration(method string, path string, statusCode int, duration time.Duration)
	ObserveRequestBodySize(method string, path string, size int)
	ObserveResponseBodySize(method string, path string, size int)
}

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

func Metrics(storage MetricStorage) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var scSrc scSource
			buf, isBuffer := w.(*buffer.Buffer)
			if isBuffer {
				scSrc = buf
			} else {
				scSrc = &writerWrapper{ResponseWriter: w}
			}

			start := time.Now()
			err := next(ctx, w, r)
			storage.ObserveDuration(r.Method, r.URL.Path, scSrc.StatusCode(), time.Since(start))
			if isBuffer {
				storage.ObserveRequestBodySize(r.Method, r.URL.Path, len(buf.RequestBody()))
				storage.ObserveResponseBodySize(r.Method, r.URL.Path, len(buf.ResponseBody()))
			}

			return err
		}
	}
}
