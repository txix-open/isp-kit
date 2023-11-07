package endpoint

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/integration-system/isp-kit/http/apierrors"
	"github.com/integration-system/isp-kit/http/endpoint/buffer"
	"github.com/integration-system/isp-kit/log/logutil"

	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/pkg/errors"
)

const (
	requestIdHeader           = "x-request-id"
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

func MaxRequestBodySize(maxBytes int64) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			return next(ctx, w, r)
		}
	}
}

func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
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

func ErrorHandler(logger log.Logger) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := next(ctx, w, r)
			if err == nil {
				return nil
			}

			logFunc := logutil.LogLevelFuncForError(err, logger)
			logFunc(ctx, err)

			var httpErr HttpError
			if errors.As(err, &httpErr) {
				err = httpErr.WriteError(w)
				return err
			}

			//hide error details to prevent potential security leaks
			err = apierrors.NewInternalServiceError(err).WriteError(w)

			return err
		}
	}
}

func RequestId() Middleware {
	return func(next HandlerFunc) HandlerFunc {
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

func Log(logger log.Logger, availableContentTypes []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
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

func DefaultLog(logger log.Logger) Middleware {
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
