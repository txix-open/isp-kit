package log

import (
	"context"
)

type contextKey struct{}

// nolint:gochecknoglobals
var (
	contextKeyValue = contextKey{}
)

// ContextLogValues extracts logging fields from a context.
func ContextLogValues(ctx context.Context) []Field {
	value, _ := ctx.Value(contextKeyValue).([]Field)
	return value
}

// ToContext adds logging fields to a context.
func ToContext(ctx context.Context, kvs ...Field) context.Context {
	existedValues := append(ContextLogValues(ctx), kvs...)
	return context.WithValue(ctx, contextKeyValue, existedValues)
}

// RewriteContextField replaces a field in the context if it exists.
func RewriteContextField(ctx context.Context, field Field) context.Context {
	existedValues := ContextLogValues(ctx)
	fields := make([]Field, 0, len(existedValues))

	for i := range existedValues {
		if existedValues[i].Key == field.Key {
			fields = append(fields, field)
		} else {
			fields = append(fields, existedValues[i])
		}
	}

	return context.WithValue(ctx, contextKeyValue, fields)
}

// UpsertContextField replaces a field if it exists, or adds it if it doesn't.
func UpsertContextField(ctx context.Context, field Field) context.Context {
	var found bool
	existedValues := ContextLogValues(ctx)
	fields := make([]Field, 0, len(existedValues))

	for i := range existedValues {
		if existedValues[i].Key == field.Key {
			fields = append(fields, field)
			found = true
		} else {
			fields = append(fields, existedValues[i])
		}
	}

	if !found {
		fields = append(fields, field)
	}
	return context.WithValue(ctx, contextKeyValue, fields)
}

// FromContext retrieves a field from the context by key.
func FromContext(ctx context.Context, key string) (Field, bool) {
	existedValues := ContextLogValues(ctx)

	for i := range existedValues {
		if existedValues[i].Key == key {
			return existedValues[i], true
		}
	}

	return Field{}, false
}

// CopyValues copies logging fields from one context to another.
func CopyValues(ctxTo context.Context, ctxFrom context.Context) context.Context {
	return ToContext(ctxTo, ContextLogValues(ctxFrom)...)
}
