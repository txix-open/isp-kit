package log

import (
	"context"
)

type Logger interface {
	Error(ctx context.Context, message any, fields ...Field)
	Warn(ctx context.Context, message any, fields ...Field)
	Info(ctx context.Context, message any, fields ...Field)
	Debug(ctx context.Context, message any, fields ...Field)
}
