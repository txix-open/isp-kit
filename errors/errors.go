// Package errors provides helpers for creating, wrapping, and inspecting errors
// using the standard library errors and fmt packages.
package errors

import (
	stderrors "errors"
	"fmt"
)

// ErrUnsupported indicates that an operation is not supported.
var ErrUnsupported = stderrors.ErrUnsupported

// New returns an error with the supplied text.
func New(text string) error {
	return stderrors.New(text) //nolint:err113
}

// Errorf formats according to a format specifier and returns an error.
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...) //nolint:err113
}

// WithMessage wraps err with message using standard error wrapping.
func WithMessage(err error, message string) error {
	return Errorf("%s: %w", message, err)
}

// WithMessagef wraps err with a formatted message using standard error wrapping.
func WithMessagef(err error, format string, args ...any) error {
	return Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// Is reports whether err's tree contains target.
func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

// As finds the first error in err's tree that matches target.
func As(err error, target any) bool {
	return stderrors.As(err, target)
}

// AsType finds the first error in err's tree that matches E.
func AsType[E error](err error) (E, bool) {
	return stderrors.AsType[E](err)
}

// Unwrap returns the result of calling err's Unwrap method.
func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

// Join returns an error that wraps the given errors.
func Join(errs ...error) error {
	return stderrors.Join(errs...)
}
