package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/txix-open/isp-kit/log"
)

// NoopHub is a no-operation implementation of the Hub interface.
// It discards all events and errors without sending them to Sentry.
// Use this when Sentry is disabled to avoid nil pointer issues.
// The NoopHub is safe for concurrent use.
type NoopHub struct {
}

// NewNoopHub creates a new NoopHub instance.
func NewNoopHub() NoopHub {
	return NoopHub{}
}

// CatchError discards the error without any action.
func (n NoopHub) CatchError(ctx context.Context, err error, level log.Level) {
}

// CatchEvent discards the event without any action.
func (n NoopHub) CatchEvent(ctx context.Context, event *sentry.Event) {

}

// Flush performs no action.
func (n NoopHub) Flush() {

}
