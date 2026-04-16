// Package apierrors provides HTTP error handling with structured error responses.
// It supports business and internal errors with customizable HTTP status codes,
// error codes, and logging levels.
package apierrors

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

const (
	// ErrCodeInternal is the default error code for internal service errors.
	ErrCodeInternal = 900
)

// Error represents a structured HTTP error with an error code, message, and optional details.
// It supports custom HTTP status codes, logging levels, and error chaining.
type Error struct {
	ErrorCode    int
	ErrorMessage string
	Details      map[string]any `json:",omitempty"`

	httpStatusCode int
	cause          error
	level          log.Level
}

// NewInternalServiceError creates a new internal service error with HTTP 500 status code.
// It logs at Error level and wraps the provided cause error.
func NewInternalServiceError(err error) Error {
	return New(http.StatusInternalServerError, ErrCodeInternal, "internal service error", err).WithLogLevel(log.ErrorLevel)
}

// NewBusinessError creates a new business error with HTTP 400 status code.
// It logs at Warn level and is used for client-side validation or business logic errors.
func NewBusinessError(errorCode int, errorMessage string, err error) Error {
	return New(http.StatusBadRequest, errorCode, errorMessage, err).WithLogLevel(log.WarnLevel)
}

// New creates a new Error with the specified HTTP status code, error code, message, and cause.
// The default log level is ErrorLevel.
func New(
	httpStatusCode int,
	errorCode int,
	errorMessage string,
	err error,
) Error {
	return Error{
		ErrorCode:      errorCode,
		ErrorMessage:   errorMessage,
		httpStatusCode: httpStatusCode,
		cause:          err,
		level:          log.ErrorLevel,
	}
}

// Error returns a formatted string representation of the error.
func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.ErrorCode, e.ErrorMessage, e.cause)
}

// WriteError writes the error as a JSON response to the http.ResponseWriter.
// It sets the Content-Type header to application/json and the appropriate HTTP status code.
func (e Error) WriteError(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.httpStatusCode)
	err := json.NewEncoder(w).Encode(e)
	if err != nil {
		return errors.WithMessage(err, "json encode error")
	}
	return nil
}

// WithDetails adds additional context details to the error.
func (e Error) WithDetails(details map[string]any) Error {
	e.Details = details
	return e
}

// WithLogLevel sets the logging level for this error.
func (e Error) WithLogLevel(level log.Level) Error {
	e.level = level
	return e
}

// LogLevel returns the logging level associated with this error.
func (e Error) LogLevel() log.Level {
	return e.level
}
