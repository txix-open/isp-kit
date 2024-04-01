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
	err := json.NewDecoder(reader).Decode(instance.Interface())
	if err != nil {
		err = errors.WithMessage(err, "unmarshal json request body")
		return reflect.Value{}, apierrors.NewBusinessError(http.StatusBadRequest, err.Error(), err)
	}

	elem := instance.Elem()

	ok, details := j.Validator.Validate(elem.Interface())
	if ok {
		return elem, nil
	}
	formattedDetails := formatDetails(details)
	return reflect.Value{}, apierrors.NewBusinessError(
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
