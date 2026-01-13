package httplog

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/buffer"
	"github.com/txix-open/isp-kit/log"
)

const (
	applicationNameHeader = "X-Application-Name"
	applicationIdHeader   = "X-Application-Identity"
)

// nolint:gochecknoglobals
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
	combinedLog         bool
}

func Log(logger log.Logger, logBody bool) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      logBody,
		logResponseBody:     logBody,
	}
	return middleware(logger, cfg)
}

func CombinedLog(logger log.Logger, logBody bool) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      logBody,
		logResponseBody:     logBody,
	}
	return combinedLogMiddleware(logger, cfg)
}

func LogWithOptions(logger log.Logger, opts ...Option) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      false,
		logResponseBody:     false,
		combinedLog:         false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.combinedLog {
		return combinedLogMiddleware(logger, cfg)
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

func combinedLogMiddleware(logger log.Logger, cfg *logConfig) endpoint.LogMiddleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			buf := buffer.Acquire(w)
			defer buffer.Release(buf)

			now := time.Now()
			logFields := []log.Field{
				log.String("method", r.Method),
				log.String("url", r.URL.String()),
			}
			logFields = append(logFields, applicationLogFields(r)...)

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

				logFields = append(logFields, log.ByteString("requestBody", buf.RequestBody()))
			}

			err := next(ctx, buf, r)

			logFields = append(logFields,
				log.Int("statusCode", buf.StatusCode()),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			)
			responseContentType := buf.Header().Get("Content-Type")
			if cfg.logResponseBody && matchContentType(responseContentType, cfg.logBodyContentTypes) {
				logFields = append(logFields, log.ByteString("responseBody", buf.ResponseBody()))
			}

			logger.Debug(ctx, "http handler: log", logFields...)

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

func applicationLogFields(r *http.Request) []log.Field {
	logFields := make([]log.Field, 0)
	appName := r.Header.Get(applicationNameHeader)
	if appName != "" {
		logFields = append(logFields, log.String("applicationName", appName))
	}

	appId := r.Header.Get(applicationIdHeader)
	if appId == "" {
		return logFields
	}

	intAppId, err := strconv.Atoi(appId)
	if err == nil {
		logFields = append(logFields, log.Int("applicationId", intAppId))
	}
	return logFields
}
