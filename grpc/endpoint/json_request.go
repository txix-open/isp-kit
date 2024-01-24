package endpoint

import (
	"context"
	"reflect"
	"unicode"

	"github.com/integration-system/isp-kit/grpc/apierrors"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
)

const (
	errCodeBadRequest = 400
)

type Validator interface {
	Validate(value any) (bool, map[string]string)
}

type JsonRequestExtractor struct {
	Validator Validator
}

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

func formatDetails(details map[string]string) map[string]any {
	result := make(map[string]any, len(details))
	for k, v := range details {
		arr := []rune(k)
		arr[0] = unicode.ToLower(arr[0])
		result[string(arr)] = v
	}
	return result
}
