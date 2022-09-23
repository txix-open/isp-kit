package httpcli

import (
	"context"
	"net/http"
)

type RequestBuilder struct {
	method        string
	url           string
	headers       map[string]string
	cookies       []*http.Cookie
	requestBody   RequestBodyWriter
	responseBody  ResponseBodyReader
	multipartData *MultipartData
	basicAuth     *BasicAuth
	queryParams   map[string]any
	retryOptions  *retryOptions

	execute func(ctx context.Context, req *RequestBuilder) (*Response, error)
}

func NewRequestBuilder(method string, url string, cfg GlobalRequestConfig, execute func(ctx context.Context, req *RequestBuilder) (*Response, error)) *RequestBuilder {
	builder := &RequestBuilder{
		method:       method,
		url:          url,
		execute:      execute,
		retryOptions: noRetries,
	}
	cfg.configure(builder)
	return builder
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

func (b *RequestBuilder) Do(ctx context.Context) (*Response, error) {
	resp, err := b.execute(ctx, b)
	b.execute = nil
	return resp, err
}
