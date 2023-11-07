package sentry_http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/integration-system/isp-kit/http/endpoint"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/log/logutil"
	"github.com/integration-system/isp-kit/requestid"
)

type Hub interface {
	CatchEvent(ctx context.Context, event *sentry.Event)
}

func Middleware(hub Hub) endpoint.Middleware {
	return func(next endpoint.HandlerFunc) endpoint.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			err := next(ctx, w, r)
			if err == nil {
				return nil
			}

			level := logutil.LogLevelForError(err)
			if level != log.ErrorLevel {
				return err
			}

			event := &sentry.Event{
				Extra: map[string]interface{}{
					"requestId": requestid.FromContext(ctx),
				},
				Level:     sentry.LevelError,
				Message:   err.Error(),
				Timestamp: time.Now(),
				Request:   request(r),
			}
			event.SetException(err, 10)
			hub.CatchEvent(ctx, event)

			return err
		}
	}
}

func request(r *http.Request) *sentry.Request {
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
