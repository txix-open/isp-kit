package endpoint

import (
	"context"
	"io"
	"net/http"
	"reflect"

	"github.com/integration-system/isp-kit/http/httperrors"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
)

type Validator interface {
	ValidateToError(v interface{}) error
}

type JsonRequestExtractor struct {
	Validator Validator
}

func (j JsonRequestExtractor) Extract(ctx context.Context, reader io.ReadCloser, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)
	err := json.NewDecoder(reader).Decode(instance.Interface())
	if err != nil {
		return reflect.Value{}, httperrors.NewHttpError(http.StatusBadRequest, errors.Errorf("unmarshal request body: %v", err))
	}

	elem := instance.Elem()

	err = j.Validator.ValidateToError(elem.Interface())
	if err != nil {
		return reflect.Value{}, httperrors.NewHttpError(http.StatusBadRequest, err)
	}

	return elem, nil
}
