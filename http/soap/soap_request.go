package soap

import (
	"context"
	"encoding/xml"
	"io"
	"reflect"

	"github.com/pkg/errors"
)

// Validator validates a value and returns an error if validation fails.
type Validator interface {
	ValidateToError(v any) error
}

// RequestExtractor extracts and validates SOAP request bodies from XML.
// It decodes the SOAP envelope, extracts the body content, and validates it.
type RequestExtractor struct {
	Validator Validator
}

// Extract decodes the SOAP envelope from the reader and extracts the body content
// into a reflect.Value of the specified type. It validates the decoded value and
// returns a SOAP fault if validation fails.
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

// ExtractV2 decodes the SOAP envelope from the reader and extracts the body content
// into the provided pointer. It validates the decoded value and returns a SOAP fault
// if validation fails.
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

// parseEnvelope decodes the SOAP envelope from the XML reader.
// It wraps the target content in an envelope structure for proper XML decoding.
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
