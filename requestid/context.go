// Package requestid provides utilities for managing request IDs in Go applications.
// It supports storing and retrieving request IDs from context, along with constants
// for common usage patterns like HTTP headers and log keys.
package requestid

import (
	"context"
)

const (
	// Header is the standard HTTP header key for request IDs.
	Header = "x-request-id"
	// LogKey is the key used for request IDs in structured logging.
	LogKey = "requestId"
)

type contextKey struct{}

// nolint:gochecknoglobals
var (
	contextKeyValue = contextKey{}
)

// ToContext stores the request ID in the context and returns the derived context.
// The request ID can be retrieved later using FromContext.
func ToContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, contextKeyValue, value)
}

// FromContext extracts the request ID from the context.
// Returns an empty string if no request ID is set in the context.
func FromContext(ctx context.Context) string {
	value, _ := ctx.Value(contextKeyValue).(string)
	return value
}
