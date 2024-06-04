package endpoint

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	http2 "gitlab.txix.ru/isp/isp-kit/http"
	"gitlab.txix.ru/isp/isp-kit/http/apierrors"
	"gitlab.txix.ru/isp/isp-kit/http/endpoint/buffer"
	"gitlab.txix.ru/isp/isp-kit/log/logutil"
	sentry2 "gitlab.txix.ru/isp/isp-kit/observability/sentry"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/requestid"
)

const (
	requestIdHeader           = "x-request-id"
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

func MaxRequestBodySize(maxBytes int64) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return next(ctx, w, r)
		}
	}
}

func Recovery() http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				recovered, ok := r.(error)
				if ok {
					err = recovered
				} else {
					err = fmt.Errorf("%v", recovered)
				}
				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, false)
				err = errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length])
			}()

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
			requestId := r.Header.Get(requestIdHeader)
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))

			return next(ctx, w, r)
		}
	}
}

var defaultAvailableContentTypes = []string{
	"application/json",
	`application/json; charset="utf-8"`,
	"text/xml",
	`text/xml; charset="utf-8"`,
}

func Log(logger log.Logger, availableContentTypes []string) http2.Middleware {
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
			if matchContentType(requestContentType, availableContentTypes) {
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
			if matchContentType(responseContentType, availableContentTypes) {
				responseLogFields = append(responseLogFields, log.ByteString("responseBody", buf.ResponseBody()))
			}

			logger.Debug(ctx, "http handler: response", responseLogFields...)

			return err
		}
	}
}

func DefaultLog(logger log.Logger) http2.Middleware {
	return Log(logger, defaultAvailableContentTypes)
}

func matchContentType(contentType string, availableContentTypes []string) bool {
	for _, content := range availableContentTypes {
		if strings.HasPrefix(contentType, content) {
			return true
		}
	}
	return false
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
