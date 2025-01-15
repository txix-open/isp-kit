package httpcli

import (
	"context"
	"net/http"
	"strings"
	"time"
)

type RequestBuilder struct {
	method            string
	url               string
	baseUrl           string
	headers           map[string]string
	cookies           []*http.Cookie
	requestBody       RequestBodyWriter
	responseBody      ResponseBodyReader
	multipartData     *MultipartData
	basicAuth         *BasicAuth
	queryParams       map[string]any
	retryOptions      *retryOptions
	timeout           time.Duration
	statusCodeToError bool
	middlewares       []Middleware

	execute func(ctx context.Context, req *RequestBuilder) (*Response, error)
}

func NewRequestBuilder(method string, url string, cfg *GlobalRequestConfig, execute func(ctx context.Context, req *RequestBuilder) (*Response, error)) *RequestBuilder {
	builder := &RequestBuilder{
		method:       method,
		url:          url,
		execute:      execute,
		retryOptions: noRetries,
	}
	cfg.configure(builder)
	return builder
}

func (b *RequestBuilder) BaseUrl(baseUrl string) *RequestBuilder {
	b.baseUrl = baseUrl
	return b
}

func (b *RequestBuilder) Url(url string) *RequestBuilder {
	b.url = url
	return b
}

func (b *RequestBuilder) Method(method string) *RequestBuilder {
	b.method = method
	return b
}

func (b *RequestBuilder) Header(name string, value string) *RequestBuilder {
	if b.headers == nil {
		b.headers = map[string]string{}
	}
	b.headers[name] = value
	return b
}

func (b *RequestBuilder) Cookie(cookie *http.Cookie) *RequestBuilder {
	b.cookies = append(b.cookies, cookie)
	return b
}

func (b *RequestBuilder) RequestBody(body []byte) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = plainRequest{value: body}
	return b
}

func (b *RequestBuilder) JsonRequestBody(value any) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = jsonRequest{value: value}
	return b
}

// JsonResponseBody
// If response status code between 200 and 299, unmarshal response body to responsePtr
func (b *RequestBuilder) JsonResponseBody(responsePtr any) *RequestBuilder {
	b.responseBody = jsonResponse{ptr: responsePtr}
	return b
}

func (b *RequestBuilder) FormDataRequestBody(data map[string][]string) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = formRequest{data: data}
	return b
}

func (b *RequestBuilder) BasicAuth(ba BasicAuth) *RequestBuilder {
	b.basicAuth = &ba
	return b
}

func (b *RequestBuilder) QueryParams(queryParams map[string]any) *RequestBuilder {
	b.queryParams = queryParams
	return b
}

func (b *RequestBuilder) Retry(cond RetryCondition, retryer Retryer) *RequestBuilder {
	b.retryOptions = &retryOptions{
		condition: cond,
		retrier:   retryer,
	}
	return b
}

// MultipartRequestBody
// Useful function to transfer "big" files because it does not load files into memory
// Does not support Request.Body in middlewares and ignores Retry
func (b *RequestBuilder) MultipartRequestBody(data *MultipartData) *RequestBuilder {
	b.requestBody = nil
	b.retryOptions = noRetries
	b.multipartData = data
	return b
}

// StatusCodeToError
// If set and Response.IsSuccess is false, Do return ErrorResponse as error
func (b *RequestBuilder) StatusCodeToError() *RequestBuilder {
	b.statusCodeToError = true
	return b
}

// Timeout
// Set per request attempt timeout, default timeout 15 seconds
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.timeout = timeout
	return b
}

func (b *RequestBuilder) Middlewares(middlewares ...Middleware) *RequestBuilder {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

func (b *RequestBuilder) Do(ctx context.Context) (*Response, error) {
	resp, err := b.execute(ctx, b)
	b.execute = nil
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() && b.statusCodeToError {
		body, err := resp.BodyCopy()
		if err != nil {
			return nil, err
		}
		return nil, ErrorResponse{
			Url:        resp.Raw.Request.URL,
			StatusCode: resp.StatusCode(),
			Body:       body,
		}
	}
	return resp, err
}

func (b *RequestBuilder) DoWithoutResponse(ctx context.Context) error {
	resp, err := b.Do(ctx)
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}

func (b *RequestBuilder) DoAndReadBody(ctx context.Context) ([]byte, int, error) {
	resp, err := b.Do(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Close()

	body, err := resp.BodyCopy()
	if err != nil {
		return nil, 0, err
	}

	return body, resp.StatusCode(), nil
}

func (b *RequestBuilder) newHttpRequest(ctx context.Context) (*http.Request, error) {
	targetUrl := b.baseUrl
	if targetUrl == "" {
		targetUrl = b.url
	}

	request, err := http.NewRequestWithContext(ctx, b.method, targetUrl, nil)
	if err != nil {
		return nil, err
	}

	if b.baseUrl != "" {
		if !strings.HasSuffix(request.URL.Path, "/") {
			request.URL.Path += "/"
		}
		request.URL = request.URL.JoinPath(b.url)
	}

	return request, nil
}
