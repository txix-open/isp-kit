package app

import (
	"context"
)

// Runner represents a long-running service or task.
// Implementations should block while running and return an error on failure.
// The context is cancelled when the application shuts down.
type Runner interface {
	Run(ctx context.Context) error
}

// RunnerFunc is a function adapter that implements the Runner interface.
// It allows using simple functions as runners without defining a new type.
//
// Example usage:
//
//	run := app.RunnerFunc(func(ctx context.Context) error {
//		for {
//			select {
//			case <-ctx.Done():
//				return nil
//			// process work
//			}
//		}
//	})
type RunnerFunc func(ctx context.Context) error

// Run executes the underlying function with the provided context.
func (r RunnerFunc) Run(ctx context.Context) error {
	return r(ctx)
}
