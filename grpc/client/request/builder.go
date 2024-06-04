package request

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/grpc"
	"gitlab.txix.ru/isp/isp-kit/grpc/isp"
	"gitlab.txix.ru/isp/isp-kit/json"
	"google.golang.org/grpc/metadata"
)

type Builder struct {
	Endpoint      string
	MD            metadata.MD
	requestBody   any
	responsePtr   any
	applicationId int
	timeout       time.Duration
	roundTripper  RoundTripper
}

func NewBuilder(roundTripper RoundTripper, endpoint string) *Builder {
	return &Builder{
		Endpoint:     endpoint,
		MD:           metadata.New(make(map[string]string)),
		roundTripper: roundTripper,
		timeout:      15 * time.Second,
	}
}

func (req *Builder) ApplicationId(appId int) *Builder {
	req.applicationId = appId
	return req
}

func (req *Builder) JsonRequestBody(reqBody any) *Builder {
	req.requestBody = reqBody
	return req
}

func (req *Builder) JsonResponseBody(respPtr any) *Builder {
	req.responsePtr = respPtr
	return req
}

func (req *Builder) Timeout(timeout time.Duration) *Builder {
	req.timeout = timeout
	return req
}

func (req *Builder) AppendMetadata(k string, v ...string) *Builder {
	req.MD[k] = append(req.MD[k], v...)
	return req
}

func (req *Builder) Do(ctx context.Context) error {
	var (
		reqCtx = ctx
		cancel context.CancelFunc
	)
	if req.timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, req.timeout)
		defer cancel()
	}

	req.MD.Set(grpc.ProxyMethodNameHeader, req.Endpoint)
	if req.applicationId != 0 {
		req.MD.Set(grpc.ApplicationIdHeader, strconv.Itoa(req.applicationId))
	}
	ctx = metadata.NewOutgoingContext(reqCtx, req.MD)
	var body []byte
	var err error
	if req.requestBody != nil {
		body, err = json.Marshal(req.requestBody)
		if err != nil {
			return errors.WithMessage(err, "marshal to json request body")
		}
	}
	message := &isp.Message{Body: &isp.Message_BytesBody{BytesBody: body}}

	resp, err := req.roundTripper(ctx, req, message)
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
