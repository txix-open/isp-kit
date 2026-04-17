package endpoint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/log/logutil"
	sentry2 "github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/panic_recovery"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const (
	// defaultMaxRequestBodySize is the default maximum allowed request body size (64MB).
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

// LogMiddleware is an alias for http.Middleware used for request/response logging.
type LogMiddleware http2.Middleware

// MaxRequestBodySize limits the size of the request body that can be read.
// It wraps the request body with http.MaxBytesReader to enforce the limit.
func MaxRequestBodySize(maxBytes int64) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return next(ctx, w, r)
		}
	}
}

// Recovery is a middleware that catches panics and converts them to errors.
// It ensures the server remains stable even when handler code panics.
// Safe for concurrent use.
func Recovery() http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer panic_recovery.Recover(func(panicErr error) {
				err = panicErr
			})

			return next(ctx, w, r)
		}
	}
}

// ErrorHandler is a middleware that logs errors and writes appropriate HTTP responses.
// It uses the error's logging level and enriches Sentry events with request details.
// For HttpError implementations, it writes the error using WriteError; otherwise,
// it returns a generic internal service error to hide implementation details.
func ErrorHandler(logger log.Logger) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := next(ctx, w, r)
			if err == nil {
				return nil
			}

			logFunc := logutil.LogLevelFuncForError(err, logger)
			logContext := sentry2.EnrichEvent(ctx, func(event *sentry.Event) {
				event.Request = sentryRequest(r)
			})
			logFunc(logContext, err)

			reqId := requestid.FromContext(ctx)
			if reqId != "" {
				w.Header().Set(requestid.Header, reqId)
			}

			var httpErr HttpError
			if errors.As(err, &httpErr) {
				err = httpErr.WriteError(w)
				return err
			}

			// hide error details to prevent potential security leaks
			err = apierrors.NewInternalServiceError(err).WriteError(w)

			return err
		}
	}
}

// RequestId is a middleware that manages request IDs for tracing and logging.
// It uses the existing request ID from headers if present, or generates a new one.
// The request ID is added to the context and response headers.
func RequestId() http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			requestId := r.Header.Get(requestid.Header)
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))

			return next(ctx, w, r)
		}
	}
}

// sentryRequest creates a Sentry request object from an http.Request.
// It determines the protocol based on TLS or X-Forwarded-Proto header.
func sentryRequest(r *http.Request) *sentry.Request {
	protocol := "https"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		protocol = "http"
	}
	url := fmt.Sprintf("%s://%s%s", protocol, r.Host, r.URL.Path)

	return &sentry.Request{
		URL:         url,
		Method:      r.Method,
		QueryString: r.URL.RawQuery,
	}
}
