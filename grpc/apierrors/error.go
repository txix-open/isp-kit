// Package apierrors provides structured error handling for gRPC services.
// It defines custom error types with business error codes, gRPC status codes,
// and structured details for comprehensive error reporting.
//
// Errors can be converted to gRPC status errors with JSON-encoded details,
// enabling rich error information to be transmitted across service boundaries.
package apierrors

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// ErrCodeInternal is the default error code for internal service errors.
	ErrCodeInternal = 900
)

// Error represents a structured error with business and gRPC status codes.
// Includes optional details for field-specific errors and a configurable log level.
// Implements the GrpcError interface for conversion to gRPC status errors.
//
// Fields:
//   - ErrorCode: Business-level error code (e.g., 400 for bad request)
//   - ErrorMessage: Human-readable error message
//   - Details: Optional map of field-specific error details
//   - grpcStatusCode: gRPC status code (e.g., codes.InvalidArgument)
//   - cause: Underlying error cause
//   - level: Log level for error logging
type Error struct {
	ErrorCode    int
	ErrorMessage string
	Details      map[string]any `json:",omitempty"`

	grpcStatusCode codes.Code
	cause          error
	level          log.Level
}

// NewInternalServiceError creates a new internal service error.
// Uses gRPC status code Internal and business error code 900.
// Sets log level to Error.
func NewInternalServiceError(err error) Error {
	return New(codes.Internal, ErrCodeInternal, "internal service error", err).WithLogLevel(log.ErrorLevel)
}

// NewBusinessError creates a new business validation error.
// Uses gRPC status code InvalidArgument and the specified error code.
// Sets log level to Warn.
func NewBusinessError(errorCode int, errorMessage string, err error) Error {
	return New(codes.InvalidArgument, errorCode, errorMessage, err).WithLogLevel(log.WarnLevel)
}

// New creates a new Error with the specified parameters.
// Sets default log level to Error.
func New(
	grpcStatusCode codes.Code,
	errorCode int,
	errorMessage string,
	err error,
) Error {
	return Error{
		ErrorCode:      errorCode,
		ErrorMessage:   errorMessage,
		grpcStatusCode: grpcStatusCode,
		cause:          err,
		level:          log.ErrorLevel,
	}
}

// Error returns a string representation of the error.
// Includes error code, message, and underlying cause.
func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.ErrorCode, e.ErrorMessage, e.cause)
}

// GrpcStatusError converts the Error to a gRPC status error.
// Serializes the error to JSON and attaches it as gRPC status details.
// Returns an error if JSON marshaling or status creation fails.
func (e Error) GrpcStatusError() error {
	data, err := json.Marshal(e)
	if err != nil {
		return errors.WithMessage(err, "marshal json")
	}

	msg := &isp.Message{
		Body: &isp.Message_BytesBody{
			BytesBody: data,
		},
	}
	s, err := status.New(e.grpcStatusCode, e.ErrorMessage).WithDetails(msg)
	if err != nil {
		return errors.WithMessage(err, "set status details")
	}
	return s.Err()
}

// WithDetails sets the error details map.
// Returns the Error for method chaining.
func (e Error) WithDetails(details map[string]any) Error {
	e.Details = details
	return e
}

// WithLogLevel sets the log level for this error.
// Returns the Error for method chaining.
func (e Error) WithLogLevel(level log.Level) Error {
	e.level = level
	return e
}

// LogLevel returns the log level for this error.
func (e Error) LogLevel() log.Level {
	return e.level
}

// FromError extracts an Error from a gRPC status error.
// Returns nil if the error is not a gRPC status error or if the details cannot be parsed.
// Useful for checking if an error originated from a remote service.
func FromError(err error) *Error {
	s, ok := status.FromError(err)
	if !ok {
		return nil
	}

	for _, detail := range s.Details() {
		typedDetail, ok := detail.(*isp.Message)
		if ok {
			errData := Error{}
			err := json.Unmarshal(typedDetail.GetBytesBody(), &errData)
			// not an error
			if err != nil {
				return nil // nolint:nilerr
			}
			if errData.ErrorCode == 0 {
				return nil
			}
			return &errData
		}
	}

	return nil
}
