package requestid

import (
	"context"
)

const (
	Header = "x-request-id"
	LogKey = "requestId"
)

type contextKey struct{}

// nolint:gochecknoglobals
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
