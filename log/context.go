package log

import (
	"context"
)

type contextLogKey struct{}

var (
	contextKey = contextLogKey{}
)

func ContextLogValues(ctx context.Context) []Field {
	value, ok := ctx.Value(contextKey).([]Field)
	if ok {
		return value
	}
	return nil
}

func ToContext(ctx context.Context, kvs ...Field) context.Context {
	existedValues := append(ContextLogValues(ctx), kvs...)
	return context.WithValue(ctx, contextKey, existedValues)
}
