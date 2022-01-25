package log

import (
	"context"
)

type Logger interface {
	Info(ctx context.Context, message interface{}, fields ...Field)
	Error(ctx context.Context, message interface{}, fields ...Field)
	Debug(ctx context.Context, message interface{}, fields ...Field)
}
