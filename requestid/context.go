package requestid

import (
	"context"
)

type contextKey int

var (
	contextKeyValue = contextKey(-1)
)

func ToContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKeyValue, value)
}

func FromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextKeyValue).(string)
	return value
}
