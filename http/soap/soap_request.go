package soap

import (
	"context"
	"encoding/xml"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

type Validator interface {
	ValidateToError(v interface{}) error
}

type RequestExtractor struct {
	Validator Validator
}

func (j RequestExtractor) Extract(_ context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)

	env := Envelope{}
	err := xml.NewDecoder(reader).Decode(&env)
	if err != nil {
		return reflect.Value{}, Fault{
			Code:   "Client",
			String: errors.WithMessage(err, "xml decode envelope").Error(),
		}
	}
	err = xml.Unmarshal(env.Body.Content, instance.Interface())
	if err != nil {
		return reflect.Value{}, Fault{
			Code:   "Client",
			String: errors.WithMessage(err, "xml decode body").Error(),
		}
	}

	elem := instance.Elem()

	err = j.Validator.ValidateToError(elem.Interface())
	if err != nil {
		return reflect.Value{}, Fault{
			Code:   "Client",
			String: errors.WithMessage(err, "invalid request body").Error(),
		}
	}

	return elem, nil
}
