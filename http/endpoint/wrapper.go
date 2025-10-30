package endpoint

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"slices"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/log"
)

type HttpError interface {
	WriteError(w http.ResponseWriter) error
}

type RequestBodyExtractor interface {
	Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error)
	ExtractV2(ctx context.Context, reader io.Reader, ptr any) error
}

type ResponseBodyMapper interface {
	Map(ctx context.Context, result any, w http.ResponseWriter) error
}

type ParamBuilder func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error)

type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

type Wrapper struct {
	ParamMappers  map[string]ParamMapper
	BodyExtractor RequestBodyExtractor
	BodyMapper    ResponseBodyMapper
	Middlewares   []http2.Middleware
	Logger        log.Logger
}

func NewWrapper(
	paramMappers []ParamMapper,
	bodyExtractor RequestBodyExtractor,
	bodyMapper ResponseBodyMapper,
	logger log.Logger,
) Wrapper {
	mappers := make(map[string]ParamMapper)
	for _, mapper := range paramMappers {
		mappers[mapper.Type] = mapper
	}
	return Wrapper{
		ParamMappers:  mappers,
		BodyExtractor: bodyExtractor,
		BodyMapper:    bodyMapper,
		Logger:        logger,
	}
}

// Endpoint
// DEPRECATED
// Use reflection free version EndpointV2 with New, NewWithoutResponse, NewWithRequest, NewDefaultHttp
func (m Wrapper) Endpoint(f any) http.HandlerFunc {
	w, isWrappable := f.(Wrappable)
	if isWrappable {
		return m.EndpointV2(w)
	}

	caller, err := NewCaller(f, m.BodyExtractor, m.BodyMapper, m.ParamMappers)
	if err != nil {
		panic(err)
	}

	handler := caller.Handle
	for i := range slices.Backward(m.Middlewares) {
		handler = m.Middlewares[i](handler)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r.Context(), w, r)
		if err != nil {
			m.Logger.Error(r.Context(), err)
		}
	}
}

func (m Wrapper) EndpointV2(w Wrappable) http.HandlerFunc {
	handler := w.Wrap(m)
	for i := range slices.Backward(m.Middlewares) {
		handler = m.Middlewares[i](handler)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r.Context(), w, r)
		if err != nil {
			m.Logger.Error(r.Context(), err)
		}
	}
}

func (m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper {
	return Wrapper{
		ParamMappers:  m.ParamMappers,
		BodyExtractor: m.BodyExtractor,
		BodyMapper:    m.BodyMapper,
		Middlewares:   append(m.Middlewares, middlewares...),
		Logger:        m.Logger,
	}
}
