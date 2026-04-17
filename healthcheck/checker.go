package healthcheck

import (
	"context"
)

// Checker is the interface that wraps the basic health check method.
// Implementations should return nil if the component is healthy, or an error
// describing the problem if it is not.
type Checker interface {
	// Healthcheck performs a health check on the component.
	// It returns nil if the component is healthy, or an error otherwise.
	Healthcheck(ctx context.Context) error
}

// CheckerFunc is a function type that implements the Checker interface.
// It allows using regular functions as health check implementations.
type CheckerFunc func(ctx context.Context) error

// Healthcheck calls the underlying function f with the provided context.
func (r CheckerFunc) Healthcheck(ctx context.Context) error {
	return r(ctx)
}
