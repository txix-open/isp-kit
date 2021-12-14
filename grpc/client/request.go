package client

import (
	"context"
	"strconv"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/json"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

type RequestBuilder struct {
	requestBody   interface{}
	responsePtr   interface{}
	applicationId int
	endpoint      string
	roundTripper  RoundTripper
}

func NewRequestBuilder(roundTripper RoundTripper, endpoint string) *RequestBuilder {
	return &RequestBuilder{
		roundTripper: roundTripper,
		endpoint:     endpoint,
	}
}

func (req *RequestBuilder) ApplicationId(appId int) *RequestBuilder {
	req.applicationId = appId
	return req
}

func (req *RequestBuilder) JsonRequestBody(reqBody interface{}) *RequestBuilder {
	req.requestBody = reqBody
	return req
}

func (req *RequestBuilder) ReadJsonResponse(respPtr interface{}) *RequestBuilder {
	req.responsePtr = respPtr
	return req
}

func (req *RequestBuilder) Do(ctx context.Context) error {
	md := metadata.Pairs(grpc.ProxyMethodNameHeader, req.endpoint)
	if req.applicationId != 0 {
		md.Set(grpc.ApplicationIdHeader, strconv.Itoa(req.applicationId))
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	var body []byte
	var err error
	if req.requestBody != nil {
		body, err = json.Marshal(req.requestBody)
		if err != nil {
			return errors.WithMessage(err, "marshal to json request body")
		}
	}
	message := &isp.Message{Body: &isp.Message_BytesBody{BytesBody: body}}

	resp, err := req.roundTripper(ctx, message)
	if err != nil {
		return err
	}

	if req.responsePtr != nil {
		err = json.Unmarshal(resp.GetBytesBody(), req.responsePtr)
		if err != nil {
			return errors.WithMessage(err, "unmarshal response body")
		}
	}

	return nil
}
