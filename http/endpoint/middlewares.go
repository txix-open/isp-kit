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
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

type LogMiddleware http2.Middleware

func MaxRequestBodySize(maxBytes int64) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return next(ctx, w, r)
		}
	}
}

//nolint:nonamedreturns
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
