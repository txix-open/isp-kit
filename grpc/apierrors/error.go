package apierrors

import (
	"fmt"

	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ErrCodeInternal = 900
)

type Error struct {
	ErrorCode    int
	ErrorMessage string
	Details      map[string]interface{} `json:",omitempty"`

	grpcStatusCode codes.Code
	cause          error
	level          log.Level
}

func NewInternalServiceError(err error) Error {
	return New(codes.Internal, ErrCodeInternal, "internal service error", err).WithLogLevel(log.ErrorLevel)
}

func NewBusinessError(errorCode int, errorMessage string, err error) Error {
	return New(codes.InvalidArgument, errorCode, errorMessage, err).WithLogLevel(log.WarnLevel)
}

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

func (e Error) Error() string {
	return fmt.Sprintf("errorCode: %d, errorMessage: %s, cause: %v", e.ErrorCode, e.ErrorMessage, e.cause)
}

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

func (e Error) WithDetails(details map[string]any) Error {
	e.Details = details
	return e
}

func (e Error) WithLogLevel(level log.Level) Error {
	e.level = level
	return e
}

func (e Error) LogLevel() log.Level {
	return e.level
}

func FromError(err error) *Error {
	s, ok := status.FromError(err)
	if !ok {
		return nil
	}

	for _, detail := range s.Details() {
		switch typedDetail := detail.(type) {
		case *isp.Message:
			errData := Error{}
			err := json.Unmarshal(typedDetail.GetBytesBody(), &errData)
			if err != nil {
				return nil
			}
			if errData.ErrorCode == 0 {
				return nil
			}
			return &errData
		}
	}

	return nil
}
