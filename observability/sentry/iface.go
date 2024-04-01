package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/txix-open/isp-kit/log"
)

type Hub interface {
	CatchError(ctx context.Context, err error, level log.Level)
	CatchEvent(ctx context.Context, event *sentry.Event)
	Flush()
}
