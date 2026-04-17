// Package request provides a fluent builder for constructing and executing gRPC requests.
// It handles JSON marshaling/unmarshaling, metadata management, and timeout configuration.
package request

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/json"
	"google.golang.org/grpc/metadata"
)

const (
	// defaultTimeout is the default request timeout when none is specified.
	defaultTimeout = 15 * time.Second
)

// Builder provides a fluent API for constructing and executing gRPC requests.
// It handles JSON marshaling, metadata management, and timeout configuration.
// Builder is not safe for concurrent use; create a new instance for each request.
type Builder struct {
	Endpoint      string
	MD            metadata.MD
	requestBody   any
	responsePtr   any
	applicationId int
	timeout       time.Duration
	roundTripper  RoundTripper
}

// NewBuilder creates a new Builder for the specified endpoint.
// Initializes with default timeout and empty metadata.
func NewBuilder(roundTripper RoundTripper, endpoint string) *Builder {
	return &Builder{
		Endpoint:     endpoint,
		MD:           metadata.New(make(map[string]string)),
		roundTripper: roundTripper,
		timeout:      defaultTimeout,
	}
}

// ApplicationId sets the application identity header for the request.
// Returns the Builder for method chaining.
func (req *Builder) ApplicationId(appId int) *Builder {
	req.applicationId = appId
	return req
}

// JsonRequestBody sets the request body as a JSON-encodable value.
// Returns the Builder for method chaining.
func (req *Builder) JsonRequestBody(reqBody any) *Builder {
	req.requestBody = reqBody
	return req
}

// JsonResponseBody sets the pointer to unmarshal the JSON response into.
// Returns the Builder for method chaining.
func (req *Builder) JsonResponseBody(respPtr any) *Builder {
	req.responsePtr = respPtr
	return req
}

// Timeout sets the request timeout duration.
// A timeout of zero or negative disables the timeout.
// Returns the Builder for method chaining.
func (req *Builder) Timeout(timeout time.Duration) *Builder {
	req.timeout = timeout
	return req
}

// AppendMetadata adds one or more values to the metadata key.
// Returns the Builder for method chaining.
func (req *Builder) AppendMetadata(k string, v ...string) *Builder {
	req.MD[k] = append(req.MD[k], v...)
	return req
}

// Do executes the request and unmarshals the response if a response pointer was provided.
// Returns an error if JSON marshaling, request execution, or response unmarshaling fails.
// Automatically applies the configured timeout to the request context.
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
