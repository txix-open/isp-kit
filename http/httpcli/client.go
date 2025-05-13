package httpcli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

type RoundTripper interface {
	RoundTrip(ctx context.Context, request *Request) (*Response, error)
}

type RoundTripperFunc func(ctx context.Context, request *Request) (*Response, error)

func (f RoundTripperFunc) RoundTrip(ctx context.Context, request *Request) (*Response, error) {
	return f(ctx, request)
}

type Middleware func(next RoundTripper) RoundTripper

type Client struct {
	cli          *http.Client
	globalConfig *GlobalRequestConfig
	mws          []Middleware

	roundTripper RoundTripper
}

// nolint:mnd,gochecknoglobals
var (
	StdClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: defaultTransportDialContext(&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          512,
			MaxIdleConnsPerHost:   32,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ReadBufferSize:        8 * 1024,
			WriteBufferSize:       8 * 1024,
		},
	}
)

func New(opts ...Option) *Client {
	return NewWithClient(StdClient, opts...)
}

func NewWithClient(cli *http.Client, opts ...Option) *Client {
	c := &Client{
		cli:          cli,
		globalConfig: NewGlobalRequestConfig(),
	}
	for _, opt := range opts {
		opt(c)
	}

	c.roundTripper = joinMiddlewares(RoundTripperFunc(c.executeWithRetries), c.mws...)
	return c
}

func (c *Client) GlobalRequestConfig() *GlobalRequestConfig {
	return c.globalConfig
}

func (c *Client) Post(url string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPost, url, c.globalConfig, c.Execute)
}

func (c *Client) Get(url string) *RequestBuilder {
	return NewRequestBuilder(http.MethodGet, url, c.globalConfig, c.Execute)
}

func (c *Client) Put(url string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPut, url, c.globalConfig, c.Execute)
}

func (c *Client) Delete(url string) *RequestBuilder {
	return NewRequestBuilder(http.MethodDelete, url, c.globalConfig, c.Execute)
}

func (c *Client) Patch(url string) *RequestBuilder {
	return NewRequestBuilder(http.MethodPatch, url, c.globalConfig, c.Execute)
}

// nolint:cyclop
func (c *Client) Execute(ctx context.Context, builder *RequestBuilder) (*Response, error) {
	request, err := builder.newHttpRequest(ctx)
	if err != nil {
		return nil, err
	}

	for name, value := range builder.headers {
		request.Header.Set(name, value)
	}
	for _, cookie := range builder.cookies {
		request.AddCookie(cookie)
	}
	if builder.basicAuth != nil {
		request.SetBasicAuth(builder.basicAuth.Username, builder.basicAuth.Password)
	}
	if builder.queryParams != nil {
		values := url.Values{}
		for key, value := range builder.queryParams {
			values.Set(key, fmt.Sprintf("%v", value))
		}
		request.URL.RawQuery = values.Encode()
	}

	rr := &Request{
		Raw:          request,
		timeout:      builder.timeout,
		retryOptions: builder.retryOptions,
	}

	if builder.multipartData != nil {
		ct, reader := builder.multipartData.openReader()
		request.Header.Set("Content-Type", ct)
		request.Body = reader
	}

	if builder.requestBody != nil {
		buff := acquireBuffer()
		defer releaseBuffer(buff)
		err := builder.requestBody.Write(request, buff)
		if err != nil {
			return nil, err
		}
		rr.body = buff.Bytes()
	}

	roundTripper := c.roundTripper
	if len(builder.middlewares) > 0 {
		roundTripper = joinMiddlewares(roundTripper, builder.middlewares...)
	}

	resp, err := roundTripper.RoundTrip(ctx, rr)
	if err != nil {
		return nil, err
	}

	if resp.IsSuccess() && builder.responseBody != nil {
		body, err := resp.Body()
		if err != nil {
			return nil, err
		}
		err = builder.responseBody.Read(bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// nolint:bodyclose
func (c *Client) executeWithRetries(ctx context.Context, request *Request) (*Response, error) {
	var (
		response *Response
		err      error
	)
	origCtx := ctx
	_ = request.retryOptions.retrier.Do(ctx, func() error {
		if response != nil {
			response.Close() // prevent context and buffer leak from previous failed attempt
		}
		var (
			ctx    context.Context
			cancel context.CancelFunc
		)
		if request.timeout > 0 {
			ctx, cancel = context.WithTimeout(origCtx, request.timeout)
			request.Raw = request.Raw.WithContext(ctx)
		}
		if request.body != nil { // it's a none multipart
			request.Raw.Body = io.NopCloser(bytes.NewBuffer(request.body))
		}
		var resp *http.Response
		resp, err = c.cli.Do(request.Raw)
		buff := acquireBuffer()
		response = &Response{
			Raw:    resp,
			buff:   buff,
			cancel: cancel,
		}
		retryErr := request.retryOptions.condition(err, response)
		return retryErr
	})

	return response, err
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

// nolint:ireturn
func joinMiddlewares(root RoundTripper, mws ...Middleware) RoundTripper {
	roundTripper := root
	for i := len(mws) - 1; i >= 0; i-- {
		roundTripper = mws[i](roundTripper)
	}
	return roundTripper
}
