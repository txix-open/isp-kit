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

func RewriteContextField(ctx context.Context, field Field) context.Context {
	existedValues := ContextLogValues(ctx)
	for i := range existedValues {
		if existedValues[i].Key == field.Key {
			existedValues[i] = field
		}
	}
	return context.WithValue(ctx, contextKeyValue, existedValues)
}
