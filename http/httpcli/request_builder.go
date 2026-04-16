package httpcli

import (
	"context"
	"net/http"
	"strings"
	"time"
)

// RequestBuilder provides a fluent API for constructing HTTP requests.
// It supports method chaining for setting various request options.
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

// NewRequestBuilder creates a new RequestBuilder with the given HTTP method, URL,
// global configuration, and execute function.
func NewRequestBuilder(
	method string,
	url string,
	cfg *GlobalRequestConfig,
	execute func(ctx context.Context, req *RequestBuilder) (*Response, error),
) *RequestBuilder {
	builder := &RequestBuilder{
		method:       method,
		url:          url,
		execute:      execute,
		retryOptions: noRetries,
	}
	cfg.configure(builder)
	return builder
}

// BaseUrl sets the base URL that will be prepended to the request URL.
// Useful for API clients with a fixed base path.
func (b *RequestBuilder) BaseUrl(baseUrl string) *RequestBuilder {
	b.baseUrl = baseUrl
	return b
}

// Url sets the request URL, overriding any previously set URL.
func (b *RequestBuilder) Url(url string) *RequestBuilder {
	b.url = url
	return b
}

// Method sets the HTTP method for the request.
func (b *RequestBuilder) Method(method string) *RequestBuilder {
	b.method = method
	return b
}

// Header sets a request header with the given name and value.
func (b *RequestBuilder) Header(name string, value string) *RequestBuilder {
	if b.headers == nil {
		b.headers = map[string]string{}
	}
	b.headers[name] = value
	return b
}

// Cookie adds a cookie to the request.
func (b *RequestBuilder) Cookie(cookie *http.Cookie) *RequestBuilder {
	b.cookies = append(b.cookies, cookie)
	return b
}

// RequestBody sets the request body from a byte slice.
func (b *RequestBuilder) RequestBody(body []byte) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = plainRequest{value: body}
	return b
}

// JsonRequestBody sets the request body as JSON from the given value.
// Sets Content-Type to application/json.
func (b *RequestBuilder) JsonRequestBody(value any) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = jsonRequest{value: value}
	return b
}

// JsonResponseBody configures the builder to unmarshal the JSON response body
// into the provided pointer if the status code is in the 2xx range.
func (b *RequestBuilder) JsonResponseBody(responsePtr any) *RequestBuilder {
	b.responseBody = jsonResponse{ptr: responsePtr}
	return b
}

// FormDataRequestBody sets the request body as form-encoded data.
// Sets Content-Type to application/x-www-form-urlencoded.
func (b *RequestBuilder) FormDataRequestBody(data map[string][]string) *RequestBuilder {
	b.multipartData = nil
	b.requestBody = formRequest{data: data}
	return b
}

// BasicAuth sets HTTP basic authentication credentials for the request.
func (b *RequestBuilder) BasicAuth(ba BasicAuth) *RequestBuilder {
	b.basicAuth = &ba
	return b
}

// QueryParams sets query parameters for the request URL.
// Values are converted to strings using fmt.Sprintf.
func (b *RequestBuilder) QueryParams(queryParams map[string]any) *RequestBuilder {
	b.queryParams = queryParams
	return b
}

// Retry configures retry behavior for the request.
// The condition determines when to retry, and the retryer controls the retry logic.
func (b *RequestBuilder) Retry(cond RetryCondition, retryer Retryer) *RequestBuilder {
	b.retryOptions = &retryOptions{
		condition: cond,
		retrier:   retryer,
	}
	return b
}

// MultipartRequestBody sets up multipart/form-data request with files and values.
//
// This is useful for uploading large files as data is streamed and not buffered in memory.
//
// Note: This disables retry support and request body access in middlewares.
func (b *RequestBuilder) MultipartRequestBody(data *MultipartData) *RequestBuilder {
	b.requestBody = nil
	b.retryOptions = noRetries
	b.multipartData = data
	return b
}

// StatusCodeToError enables automatic error conversion for non-success status codes.
//
// When enabled and the response status is not in the 2xx range, Do() returns
// an ErrorResponse containing the status code and body.
func (b *RequestBuilder) StatusCodeToError() *RequestBuilder {
	b.statusCodeToError = true
	return b
}

// Timeout sets the timeout duration for the request attempt.
//
// Defaults to 15 seconds if not specified.
func (b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	b.timeout = timeout
	return b
}

// Middlewares adds one or more middlewares to be executed for this specific request.
// These are executed after the client's global middlewares.
func (b *RequestBuilder) Middlewares(middlewares ...Middleware) *RequestBuilder {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

// Do executes the request and returns the response.
//
// Returns the response and any error that occurred during execution.
// If StatusCodeToError was called and the response is not successful,
// returns ErrorResponse as the error.
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

// DoWithoutResponse executes the request and discards the response.
//
// Returns any error that occurred during execution.
// Automatically closes the response to release resources.
func (b *RequestBuilder) DoWithoutResponse(ctx context.Context) error {
	resp, err := b.Do(ctx)
	if err != nil {
		return err
	}
	resp.Close()
	return nil
}

// DoAndReadBody executes the request and returns the response body.
//
// Returns the response body, status code, and any error.
// Automatically closes the response to release resources.
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

// newHttpRequest creates an http.Request from the builder configuration.
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
