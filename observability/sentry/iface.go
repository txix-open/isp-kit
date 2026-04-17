package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/txix-open/isp-kit/log"
)

// Hub defines the interface for Sentry operations.
// It provides methods for capturing errors and events, and flushing pending events.
type Hub interface {
	// CatchError captures an error with the specified log level.
	// The error is extracted from the context and sent to Sentry.
	CatchError(ctx context.Context, err error, level log.Level)
	// CatchEvent captures a Sentry event directly.
	CatchEvent(ctx context.Context, event *sentry.Event)
	// Flush waits for any buffered events to be sent to Sentry.
	Flush()
}
