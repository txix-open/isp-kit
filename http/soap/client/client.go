package client

import (
	"context"
	"encoding/xml"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/http/soap"
	"github.com/pkg/errors"
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
	requestBody RequestBody,
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
		data, err := requestBody.Body()
		if err != nil {
			return nil, errors.WithMessage(err, "xml marshal request body")
		}
		reqEnv := soap.Envelope{Body: soap.Body{Content: data}}
		reqBody, err := xml.Marshal(reqEnv)
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
