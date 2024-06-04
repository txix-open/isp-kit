package client

import (
	"encoding/xml"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/http/soap"
)

type Response struct {
	Http *httpcli.Response
}

func (r *Response) Close() {
	r.Http.Close()
}

func (r *Response) UnmarshalPayload(responseBody any) error {
	resBody, err := r.Http.Body()
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

func (r *Response) Fault() (*soap.Fault, error) {
	resBody, err := r.Http.Body()
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
