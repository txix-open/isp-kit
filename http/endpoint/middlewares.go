package endpoint

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/integration-system/isp-kit/http/httperrors"

	"github.com/integration-system/isp-kit/http/endpoint/writer"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
	"github.com/pkg/errors"
)

const requestIdHeader = "x-request-id"

func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}

				recovered, ok := r.(error)
				if !ok {
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

			logger.Error(ctx, err)

			httpErr, ok := err.(HttpError)
			if ok {
				_ = httpErr.WriteError(w)
				return nil
			}

			httpErr = httperrors.New(http.StatusInternalServerError, err)
			_ = httpErr.WriteError(w)

			return nil
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

func RequestInfo() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = log.ToContext(
				ctx,
				log.String("httpMethod", r.Method),
				log.String("httpUrl", r.URL.String()),
			)

			return next(ctx, w, r)
		}
	}
}

var defaultAvailableContentTypes = []string{
	"application/json",
	"application/xml",
	"text/html",
}

func BodyLogger(logger log.Logger, availableContentTypes []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			requestContentType := r.Header.Get("Content-Type")
			if matchContentType(requestContentType, availableContentTypes) {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					return errors.WithMessage(err, "read request body for logging")
				}
				err = r.Body.Close()
				if err != nil {
					return errors.WithMessage(err, "close request reader")
				}
				r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

				logger.Debug(ctx, "request body", log.ByteString("requestBody", bodyBytes))
			}

			mw := writer.Acquire(w)
			defer writer.Release(mw)

			err := next(ctx, mw, r)

			responseContentType := mw.Header().Get("Content-Type")
			if matchContentType(responseContentType, availableContentTypes) {
				logger.Debug(ctx, "response body", log.ByteString("responseBody", mw.GetBody()))
			}

			return err
		}
	}
}

func DefaultBodyLogger(logger log.Logger) Middleware {
	return BodyLogger(logger, defaultAvailableContentTypes)
}

func matchContentType(contentType string, availableContentTypes []string) bool {
	for _, content := range availableContentTypes {
		if strings.HasPrefix(contentType, content) {
			return true
		}
	}
	return false
}
