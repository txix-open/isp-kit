package endpoint

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"unicode"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/json"
)

// Validator validates a value and returns whether it's valid along with field-level error details.
type Validator interface {
	// Validate returns true if valid and a map of field names to error messages.
	Validate(value any) (bool, map[string]string)
}

// JsonRequestExtractor extracts and validates JSON request bodies.
// It decodes JSON, validates the result, and returns business errors for invalid input.
type JsonRequestExtractor struct {
	Validator Validator
}

// Extract decodes the JSON request body into a reflect.Value of the specified type.
// It validates the decoded value and returns a business error if validation fails.
func (j JsonRequestExtractor) Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := j.extract(reader, instance.Interface())
	if err != nil {
		return reflect.Value{}, err
	}

	elem := instance.Elem()
	err = j.validate(elem.Interface())
	if err != nil {
		return reflect.Value{}, err
	}

	return elem, nil
}

// ExtractV2 decodes the JSON request body into the provided pointer.
// It validates the decoded value and returns a business error if validation fails.
func (j JsonRequestExtractor) ExtractV2(_ context.Context, reader io.Reader, ptr any) error {
	err := j.extract(reader, ptr)
	if err != nil {
		return err
	}
	return j.validate(ptr)
}

// extract decodes JSON from the reader into the target pointer.
// It returns a business error with HTTP 400 status for invalid JSON.
func (j JsonRequestExtractor) extract(reader io.Reader, ptr any) error {
	err := json.NewDecoder(reader).Decode(ptr)
	if err != nil {
		err = errors.WithMessage(err, "unmarshal json request body")
		return apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}
	return nil
}

// validate validates the decoded value using the configured validator.
// It returns a business error with validation details if validation fails.
func (j JsonRequestExtractor) validate(v any) error {
	ok, details := j.Validator.Validate(v)
	if ok {
		return nil
	}
	formattedDetails := formatDetails(details)
	return apierrors.NewBusinessError(
		http.StatusBadRequest,
		"invalid request body",
		errors.Errorf("validation errors: %v", formattedDetails),
	).WithDetails(formattedDetails)
}

// formatDetails converts validation error details by lowercasing the first letter of each field name.
func formatDetails(details map[string]string) map[string]any {
	result := make(map[string]any, len(details))
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		result[string(arr)] = v
	}
	return result
}
