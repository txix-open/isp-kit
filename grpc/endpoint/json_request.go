package endpoint

import (
	"context"
	"reflect"
	"unicode"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/apierrors"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/json"
)

const (
	// errCodeBadRequest is the error code for invalid request body errors.
	errCodeBadRequest = 400
)

// Validator defines the interface for request validation.
// Implementations should validate a value and return whether it's valid
// along with field-specific error details.
type Validator interface {
	// Validate checks if the value is valid.
	// Returns true if valid, false otherwise, along with a map of field names to error messages.
	Validate(value any) (bool, map[string]string)
}

// JsonRequestExtractor extracts and validates JSON request bodies from gRPC messages.
// Unmarshals the message body into the target type and validates it using the configured validator.
type JsonRequestExtractor struct {
	Validator Validator
}

// Extract unmarshals the message body as JSON and validates it.
// Returns a reflect.Value of the created instance or an error if unmarshaling or validation fails.
// Business validation errors are returned as apierrors.BusinessError with code 400.
func (j JsonRequestExtractor) Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := json.Unmarshal(message.GetBytesBody(), instance.Interface())
	if err != nil {
		err = errors.WithMessage(err, "unmarshal json request body")
		return reflect.Value{}, apierrors.NewBusinessError(errCodeBadRequest, err.Error(), err)
	}

	elem := instance.Elem()

	ok, details := j.Validator.Validate(elem.Interface())
	if ok {
		return elem, nil
	}
	formattedDetails := formatDetails(details)
	return reflect.Value{}, apierrors.NewBusinessError(
		errCodeBadRequest,
		"invalid request body",
		errors.Errorf("validation errors: %v", formattedDetails),
	).WithDetails(formattedDetails)
}

// formatDetails converts a map of field error messages to a format suitable for API responses.
// Converts field names to lowercase (first character) for consistency.
func formatDetails(details map[string]string) map[string]any {
	result := make(map[string]any, len(details))
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		result[string(arr)] = v
	}
	return result
}
