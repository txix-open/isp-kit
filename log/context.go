package log

import (
	"context"
)

type contextKey struct{}

var (
	contextKeyValue = contextKey{}
)

func ContextLogValues(ctx context.Context) []Field {
	value, _ := ctx.Value(contextKeyValue).([]Field)
	return value
}

func ToContext(ctx context.Context, kvs ...Field) context.Context {
	existedValues := append(ContextLogValues(ctx), kvs...)
	return context.WithValue(ctx, contextKeyValue, existedValues)
}
