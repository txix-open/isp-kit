package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/errors"
)

var (
	errSource = stderrors.New("source error")
	errFirst  = stderrors.New("first error")
	errSecond = stderrors.New("second error")
)

type typedError struct {
	message string
}

func (e *typedError) Error() string {
	return e.message
}

func TestWithMessage(t *testing.T) {
	t.Parallel()

	err := errors.WithMessage(errSource, "additional context")

	require.EqualError(t, err, "additional context: source error")
	require.True(t, errors.Is(err, errSource))
	require.Same(t, errSource, errors.Unwrap(err))
}

func TestWithMessageNil(t *testing.T) {
	t.Parallel()

	require.Error(t, errors.WithMessage(nil, "additional context"))
	require.Error(t, errors.WithMessagef(nil, "additional %s", "context"))
}

func TestWithMessagef(t *testing.T) {
	t.Parallel()

	err := errors.WithMessagef(errSource, "additional %s", "context")

	require.EqualError(t, err, "additional context: source error")
	require.True(t, errors.Is(err, errSource))
}

func TestErrorfWrap(t *testing.T) {
	t.Parallel()

	err := errors.Errorf("additional context: %w", errSource)

	require.EqualError(t, err, "additional context: source error")
	require.True(t, errors.Is(err, errSource))
}

func TestAs(t *testing.T) {
	t.Parallel()

	sourceErr := &typedError{message: "typed error"}
	err := errors.WithMessage(sourceErr, "additional context")

	var actual *typedError
	require.True(t, errors.As(err, &actual))
	require.Same(t, sourceErr, actual)
}

func TestAsType(t *testing.T) {
	t.Parallel()

	sourceErr := &typedError{message: "typed error"}
	err := errors.WithMessage(sourceErr, "additional context")

	actual, ok := errors.AsType[*typedError](err)

	require.True(t, ok)
	require.Same(t, sourceErr, actual)
}

func TestJoin(t *testing.T) {
	t.Parallel()

	err := errors.Join(errFirst, errSecond)

	require.True(t, errors.Is(err, errFirst))
	require.True(t, errors.Is(err, errSecond))
}

func TestErrUnsupported(t *testing.T) {
	t.Parallel()

	err := errors.WithMessage(errors.ErrUnsupported, "operation")

	require.True(t, errors.Is(err, stderrors.ErrUnsupported))
}
