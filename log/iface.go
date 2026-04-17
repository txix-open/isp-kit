package log

import (
	"context"
)

// Logger defines the interface for structured logging.
type Logger interface {
	Error(ctx context.Context, message any, fields ...Field)
	Warn(ctx context.Context, message any, fields ...Field)
	Info(ctx context.Context, message any, fields ...Field)
	Debug(ctx context.Context, message any, fields ...Field)
}
