package client

import (
	"encoding/xml"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/soap"
)

// Response represents a SOAP response received from a web service.
// It provides methods for parsing the response payload and checking for faults.
type Response struct {
	Http *httpcli.Response
}

// Close closes the underlying HTTP response body.
func (r *Response) Close() {
	r.Http.Close()
}

// UnmarshalPayload decodes the SOAP response body into the provided struct.
// It handles both standard SOAP envelopes and raw XML (PlainXml) responses.
// Returns an error if the response body cannot be read or decoded.
func (r *Response) UnmarshalPayload(responseBody any) error {
	resBody, err := r.Http.UnsafeBody()
	if err != nil {
		return errors.WithMessage(err, "read response body")
	}

	switch r := responseBody.(type) {
	case *PlainXml:
		resEnv := embedEnvelope{Body: embedBody{}}
		err := xml.Unmarshal(resBody, &resEnv)
		if err != nil {
			return errors.WithMessage(err, "xml unmarshal envelope")
		}
		r.Value = resEnv.Body.Content
	default:
		resEnv := soap.Envelope{Body: soap.Body{Content: responseBody}}
		err = xml.Unmarshal(resBody, &resEnv)
		if err != nil {
			return errors.WithMessage(err, "xml unmarshal envelope")
		}
	}

	return nil
}

// Fault checks if the response contains a SOAP fault and returns it.
// It returns nil if no fault is present. Returns an error if the response
// body cannot be read or decoded.
func (r *Response) Fault() (*soap.Fault, error) {
	resBody, err := r.Http.UnsafeBody()
	if err != nil {
		return nil, errors.WithMessage(err, "read response body")
	}
	fault := soap.Fault{}
	resEnv := soap.Envelope{Body: soap.Body{Fault: &fault}}
	err = xml.Unmarshal(resBody, &resEnv)
	if err != nil {
		return nil, errors.WithMessage(err, "xml unmarshal envelope")
	}
	return &fault, nil
}
