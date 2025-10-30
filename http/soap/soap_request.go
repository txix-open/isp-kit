package soap

import (
	"context"
	"encoding/xml"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

type Validator interface {
	ValidateToError(v any) error
}

type RequestExtractor struct {
	Validator Validator
}

func (j RequestExtractor) Extract(_ context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error) {
	instance := reflect.New(reqBodyType)

	err := j.parseEnvelope(reader, instance.Interface())
	if err != nil {
		return reflect.Value{}, err
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

func (j RequestExtractor) ExtractV2(ctx context.Context, reader io.Reader, ptr any) error {
	err := j.parseEnvelope(reader, ptr)
	if err != nil {
		return err
	}

	err = j.Validator.ValidateToError(ptr)
	if err != nil {
		return Fault{
			Code:   "Client",
			String: errors.WithMessage(err, "invalid request body").Error(),
		}
	}

	return nil
}

func (j RequestExtractor) parseEnvelope(reader io.Reader, content any) error {
	env := Envelope{Body: Body{Content: content}}
	err := xml.NewDecoder(reader).Decode(&env)
	if err != nil {
		return Fault{
			Code:   "Client",
			String: errors.WithMessage(err, "xml decode envelope").Error(),
		}
	}
	return nil
}
