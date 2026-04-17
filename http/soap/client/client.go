// Package client provides a SOAP client for invoking SOAP web services.
// It handles XML envelope creation, SOAP action headers, and response parsing.
package client

import (
	"context"
	"encoding/xml"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/http/soap"
)

// Client is a SOAP client that wraps an httpcli.Client for making SOAP requests.
type Client struct {
	cli *httpcli.Client
}

// New creates a new SOAP client with the specified HTTP client.
func New(cli *httpcli.Client) Client {
	return Client{
		cli: cli,
	}
}

// Invoke sends a SOAP request to the specified URL with the given action and headers.
// It automatically wraps the request body in a SOAP envelope and parses the response.
// Returns a Response object for parsing the SOAP response or fault.
func (c Client) Invoke(
	ctx context.Context,
	url string,
	soapAction string,
	extraHeaders map[string]string,
	requestBody any,
) (*Response, error) {
	builder := c.cli.Post(url)
	builder.Header("Content-Type", soap.ContentType)
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
