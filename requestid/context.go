package requestid

import (
	"context"
)

type contextKey struct{}

var (
	contextKeyValue = contextKey{}
)

func ToContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKeyValue, value)
}

func FromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextKeyValue).(string)
	return value
}
