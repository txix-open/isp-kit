package log

import (
	"context"
)

type contextLogKey struct{}

var (
	contextKey = contextLogKey{}
)

func ContextLogValues(ctx context.Context) []Field {
	value, _ := ctx.Value(contextKey).([]Field)
	return value
}

func ToContext(ctx context.Context, kvs ...Field) context.Context {
	existedValues := append(ContextLogValues(ctx), kvs...)
	return context.WithValue(ctx, contextKey, existedValues)
}
