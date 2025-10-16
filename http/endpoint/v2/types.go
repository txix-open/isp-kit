package endpoint

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
)

type Wrappable interface {
	Wrap(wrapper Wrapper) http.HandlerFunc
}

type Basic[T, U any] func(ctx context.Context, req T) (U, error)

func New[T, U any](fn func(ctx context.Context, req T) (U, error)) Basic[T, U] { return fn }

func (fn Basic[T, U]) Wrap(wrapper Wrapper) http.HandlerFunc {
	return wrapper.Endpoint(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[T](ctx, wrapper, r)
		if err != nil {
			return err
		}

		resp, err := fn(ctx, req)
		if err != nil {
			return err
		}

		return wrapper.bodyMapper.Map(ctx, resp, w)
	})
}

type WithoutResponseBody[T any] func(ctx context.Context, req T) error

func NewWithoutResponse[T any](fn func(ctx context.Context, req T) error) WithoutResponseBody[T] {
	return fn
}

func (fn WithoutResponseBody[T]) Wrap(wrapper Wrapper) http.HandlerFunc {
	return wrapper.Endpoint(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[T](ctx, wrapper, r)
		if err != nil {
			return err
		}
		return fn(ctx, req)
	})
}

type WithRequest func(ctx context.Context, r *http.Request) error

func (fn WithRequest) Wrap(wrapper Wrapper) http.HandlerFunc {
	return wrapper.Endpoint(func(ctx context.Context, _ http.ResponseWriter, r *http.Request) error {
		return fn(ctx, r)
	})
}

type DefaultHttp http2.HandlerFunc

func (fn DefaultHttp) Wrap(wrapper Wrapper) http.HandlerFunc {
	return wrapper.Endpoint(http2.HandlerFunc(fn))
}

// nolint:ireturn
func extractBody[T any](ctx context.Context, w Wrapper, r *http.Request) (T, error) {
	var req T
	err := w.bodyExtractor.Extract(ctx, r.Body, &req)
	if err != nil {
		return *new(T), err
	}
	return req, nil
}
