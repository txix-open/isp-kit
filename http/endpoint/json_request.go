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

type Validator interface {
	Validate(value any) (bool, map[string]string)
}

type JsonRequestExtractor struct {
	Validator Validator
}

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

func (j JsonRequestExtractor) ExtractV2(_ context.Context, reader io.Reader, ptr any) error {
	err := j.extract(reader, ptr)
	if err != nil {
		return err
	}
	return j.validate(ptr)
}

func (j JsonRequestExtractor) extract(reader io.Reader, ptr any) error {
	err := json.NewDecoder(reader).Decode(ptr)
	if err != nil {
		err = errors.WithMessage(err, "unmarshal json request body")
		return apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}
	return nil
}

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

func formatDetails(details map[string]string) map[string]any {
	result := make(map[string]any, len(details))
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		result[string(arr)] = v
	}
	return result
}
