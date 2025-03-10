package httplog

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/buffer"
	"github.com/txix-open/isp-kit/log"
)

var (
	defaultLogBodyContentTypes = []string{
		"application/json",
		"text/xml",
	}
)

type logConfig struct {
	logBodyContentTypes []string
	logRequestBody      bool
	logResponseBody     bool
}

func Log(logger log.Logger, opts ...Option) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      false,
		logResponseBody:     false,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return middleware(logger, cfg)
}

func middleware(logger log.Logger, cfg *logConfig) endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			buf := buffer.Acquire(w)
			defer buffer.Release(buf)

			now := time.Now()
			requestLogFields := []log.Field{
				log.String("method", r.Method),
				log.String("url", r.URL.String()),
			}
			requestContentType := r.Header.Get("Content-Type")
			if cfg.logRequestBody && matchContentType(requestContentType, cfg.logBodyContentTypes) {
				err := buf.ReadRequestBody(r.Body)
				if err != nil {
					return errors.WithMessage(err, "read request body for logging")
				}
				err = r.Body.Close()
				if err != nil {
					return errors.WithMessage(err, "close request reader")
				}
				r.Body = buffer.NewRequestBody(buf.RequestBody())

				requestLogFields = append(requestLogFields, log.ByteString("requestBody", buf.RequestBody()))
			}

			logger.Debug(ctx, "http handler: request", requestLogFields...)

			err := next(ctx, buf, r)

			responseLogFields := []log.Field{
				log.Int("statusCode", buf.StatusCode()),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			}
			responseContentType := buf.Header().Get("Content-Type")
			if cfg.logResponseBody && matchContentType(responseContentType, cfg.logBodyContentTypes) {
				responseLogFields = append(responseLogFields, log.ByteString("responseBody", buf.ResponseBody()))
			}

			logger.Debug(ctx, "http handler: response", responseLogFields...)

			return err
		}
	}
}

func matchContentType(contentType string, availableContentTypes []string) bool {
	for _, content := range availableContentTypes {
		if strings.HasPrefix(contentType, content) {
			return true
		}
	}
	return false
}
