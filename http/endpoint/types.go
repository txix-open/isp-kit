package endpoint

import (
	"context"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
)

type Wrappable interface {
	Wrap(wrapper Wrapper) http2.HandlerFunc
}

type basic[Req any, Res any] func(ctx context.Context, req Req) (Res, error)

func New[Req any, Res any](fn func(ctx context.Context, req Req) (Res, error)) basic[Req, Res] {
	return fn
}

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

type withoutResponseBody[Req any] func(ctx context.Context, req Req) error

func NewWithoutResponse[Req any](fn func(ctx context.Context, req Req) error) withoutResponseBody[Req] {
	return fn
}

func (fn withoutResponseBody[Req]) Wrap(wrapper Wrapper) http2.HandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		req, err := extractBody[Req](ctx, wrapper, r)
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
