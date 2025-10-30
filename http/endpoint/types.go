package endpoint

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
)

type Wrappable interface {
	Wrap(wrapper Wrapper) http2.HandlerFunc
}

type basic[T, U any] func(ctx context.Context, req T) (U, error)

func New[T, U any](fn func(ctx context.Context, req T) (U, error)) basic[T, U] { return fn }

func (fn basic[T, U]) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[T](ctx, wrapper, r)
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

type withoutResponseBody[T any] func(ctx context.Context, req T) error

func NewWithoutResponse[T any](fn func(ctx context.Context, req T) error) withoutResponseBody[T] {
	return fn
}

func (fn withoutResponseBody[T]) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[T](ctx, wrapper, r)
		if err != nil {
			return err
		}
		return fn(ctx, req)
	}
}

type withRequest func(ctx context.Context, r *http.Request) error

func NewWithRequest(fn func(ctx context.Context, r *http.Request) error) withRequest {
	return fn
}

func (fn withRequest) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, _ http.ResponseWriter, r *http.Request) error {
		return fn(ctx, r)
	}
}

type defaultHttp http2.HandlerFunc

func NewDefaultHttp(fn func(ctx context.Context, w http.ResponseWriter, r *http.Request) error) defaultHttp {
	return fn
}

func (fn defaultHttp) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return http2.HandlerFunc(fn)
}

// nolint:ireturn
func extractBody[T any](ctx context.Context, w Wrapper, r *http.Request) (T, error) {
	var req T
	err := w.BodyExtractor.ExtractV2(ctx, r.Body, &req)
	if err != nil {
		return *new(T), err
	}
	return req, nil
}
