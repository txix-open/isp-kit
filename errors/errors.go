package errors

import (
	stderrors "errors"
	"fmt"
)

var ErrUnsupported = stderrors.ErrUnsupported

func New(text string) error {
	return stderrors.New(text) //nolint:err113
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...) //nolint:err113
}

func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WithMessagef(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

func Is(err, target error) bool {
	return stderrors.Is(err, target)
}

func As(err error, target any) bool {
	return stderrors.As(err, target)
}

func AsType[E error](err error) (E, bool) {
	return stderrors.AsType[E](err)
}

func Unwrap(err error) error {
	return stderrors.Unwrap(err)
}

func Join(errs ...error) error {
	return stderrors.Join(errs...)
}
