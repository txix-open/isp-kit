package client

import (
	"context"
	"encoding/xml"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/http/soap"
)

type Client struct {
	cli *httpcli.Client
}

func New(cli *httpcli.Client) Client {
	return Client{
		cli: cli,
	}
}

func (c Client) Invoke(
	ctx context.Context,
	url string,
	soapAction string,
	extraHeaders map[string]string,
	requestBody any,
) (*Response, error) {
	builder := c.cli.Post(url)
	builder.Header("content-type", soap.ContentType)
	if soapAction != "" {
		builder.Header(soap.ActionHeader, soapAction)
	}
	for name, value := range extraHeaders {
		builder.Header(name, value)
	}

	if requestBody != nil {
		var toMarshal any
		switch r := requestBody.(type) {
		case PlainXml:
			toMarshal = embedEnvelope{Body: embedBody{Content: r.Value}}
		default:
			toMarshal = soap.Envelope{Body: soap.Body{Content: requestBody}}
		}
		reqBody, err := xml.Marshal(toMarshal)
		if err != nil {
			return nil, errors.WithMessage(err, "xml marshal envelope")
		}
		builder.RequestBody(reqBody)
	}

	resp, err := builder.Do(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "http do")
	}

	return &Response{Http: resp}, nil
}
