package client

import (
	"encoding/xml"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/soap"
	"github.com/pkg/errors"
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
	resEnv := soap.Envelope{}
	err = xml.Unmarshal(resBody, &resEnv)
	if err != nil {
		return errors.WithMessage(err, "xml unmarshal envelope")
	}

	plain, ok := responseBody.(*PlainXml)
	if ok {
		plain.Value = resEnv.Body.Content
	} else {
		err = xml.Unmarshal(resEnv.Body.Content, &responseBody)
		if err != nil {
			return errors.WithMessage(err, "xml response response body")
		}
	}

	return nil
}
