package endpoint

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
)

// Wrappable is an interface for types that can be wrapped into an http.HandlerFunc.
// Implementations should handle request extraction and response mapping.
type Wrappable interface {
	Wrap(wrapper Wrapper) http2.HandlerFunc
}

// basic is a generic endpoint function type that accepts a request and returns a response.
type basic[Req any, Res any] func(ctx context.Context, req Req) (Res, error)

// New creates a new generic endpoint with request and response types.
// The returned endpoint automatically handles JSON body extraction and response mapping.
func New[Req any, Res any](fn func(ctx context.Context, req Req) (Res, error)) basic[Req, Res] {
	return fn
}

// Wrap implements Wrappable for basic endpoints.
// It extracts the request body, calls the endpoint function, and maps the response.
func (fn basic[Req, Res]) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[Req](ctx, wrapper, r)
		if err != nil {
			return err
		}

		resp, err := fn(ctx, req)
		if err != nil {
			return err
		}

		return wrapper.BodyMapper.Map(ctx, resp, w)
	}
}

// withoutResponseBody is an endpoint type that processes a request but returns no response body.
type withoutResponseBody[Req any] func(ctx context.Context, req Req) error

// NewWithoutResponse creates an endpoint that handles a request without returning a response body.
// Useful for operations like deletions or fire-and-forget actions.
func NewWithoutResponse[Req any](fn func(ctx context.Context, req Req) error) withoutResponseBody[Req] {
	return fn
}

// Wrap implements Wrappable for endpoints without response body.
func (fn withoutResponseBody[Req]) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[Req](ctx, wrapper, r)
		if err != nil {
			return err
		}
		return fn(ctx, req)
	}
}

// withRequest is an endpoint type that has direct access to the http.Request.
type withRequest func(ctx context.Context, r *http.Request) error

// NewWithRequest creates an endpoint that receives the raw http.Request.
// Use this when you need direct access to headers, query parameters, or other request details.
func NewWithRequest(fn func(ctx context.Context, r *http.Request) error) withRequest {
	return fn
}

// Wrap implements Wrappable for withRequest endpoints.
func (fn withRequest) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, _ http.ResponseWriter, r *http.Request) error {
		return fn(ctx, r)
	}
}

// defaultHttp is an endpoint type that uses the standard http.HandlerFunc signature.
type defaultHttp http2.HandlerFunc

// NewDefaultHttp creates an endpoint with direct access to ResponseWriter and Request.
// This is the most flexible option but bypasses automatic body extraction and mapping.
func NewDefaultHttp(fn func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) defaultHttp {
	return fn
}

// Wrap implements Wrappable for defaultHttp endpoints.
func (fn defaultHttp) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return http2.HandlerFunc(fn)
}

// extractBody extracts and unmarshals the request body into the target type.
// It uses the wrapper's BodyExtractor to handle the request body.
func extractBody[T any](ctx context.Context, w Wrapper, r *http.Request) (T, error) {
	var req T
	err := w.BodyExtractor.ExtractV2(ctx, r.Body, &req)
	if err != nil {
		return *new(T), err
	}
	return req, nil
}
