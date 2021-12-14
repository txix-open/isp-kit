package log

import (
	"context"
)

type contextLogKey int

var (
	contextKey = contextLogKey(-1)
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
