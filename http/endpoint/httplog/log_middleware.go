// Package httplog provides HTTP request/response logging middleware.
// It supports separate request and response body logging, combined logging,
// and content-type filtering.
package httplog

import (
	"context"
	"encoding/base64"
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
	// applicationNameHeader is the HTTP header for base64-encoded application name.
	applicationNameHeader = "X-Application-Name"
	// applicationIdHeader is the HTTP header for application identity (numeric ID).
	applicationIdHeader = "X-Application-Identity"
)

var (
	// defaultLogBodyContentTypes specifies the content types for which body logging is enabled.
	defaultLogBodyContentTypes = []string{
		"application/json",
		"text/xml",
	}
)

// logConfig holds the configuration for the logging middleware.
type logConfig struct {
	logBodyContentTypes []string
	logRequestBody      bool
	logResponseBody     bool
	combinedLog         bool
}

// Log creates a logging middleware that logs requests and responses separately.
// If logBody is true, it logs request and response bodies for supported content types.
func Log(logger log.Logger, logBody bool) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      logBody,
		logResponseBody:     logBody,
	}
	return middleware(logger, cfg)
}

// CombinedLog creates a logging middleware that logs requests and responses in a single entry.
// If logBody is true, it logs request and response bodies for supported content types.
func CombinedLog(logger log.Logger, logBody bool) endpoint.LogMiddleware {
	cfg := &logConfig{
		logBodyContentTypes: defaultLogBodyContentTypes,
		logRequestBody:      logBody,
		logResponseBody:     logBody,
	}
	return combinedLogMiddleware(logger, cfg)
}

// LogWithOptions creates a logging middleware with custom configuration options.
// By default, body logging is disabled. Use options like WithLogBody, WithCombinedLog
// to customize the behavior.
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

// middleware creates a logging middleware that logs requests and responses separately.
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

// combinedLogMiddleware creates a logging middleware that logs requests and responses in a single entry.
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

// matchContentType checks if the given content type matches any of the available content types.
// It uses prefix matching to handle content types with parameters (e.g., "application/json; charset=utf-8").
func matchContentType(contentType string, availableContentTypes []string) bool {
	for _, content := range availableContentTypes {
		if strings.HasPrefix(contentType, content) {
			return true
		}
	}
	return false
}

// applicationLogFields extracts application-related log fields from the request.
// It decodes the base64-encoded application name and parses the application ID.
func applicationLogFields(r *http.Request) []log.Field {
	logFields := make([]log.Field, 0)
	appName := r.Header.Get(applicationNameHeader)
	if appName != "" {
		decodedName, err := base64.StdEncoding.DecodeString(appName)
		if err == nil {
			logFields = append(logFields, log.String("applicationName", string(decodedName)))
		}
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
