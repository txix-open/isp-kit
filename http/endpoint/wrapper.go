package endpoint

import (
	"context"
	"io"
	"net/http"
	"reflect"

	"github.com/integration-system/isp-kit/log"
)

type HttpError interface {
	WriteError(w http.ResponseWriter) error
}

type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error
type Middleware func(next HandlerFunc) HandlerFunc

type RequestBodyExtractor interface {
	Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error)
}

type ResponseBodyMapper interface {
	Map(result interface{}, w http.ResponseWriter) error
}

type ParamBuilder func(ctx context.Context, w http.ResponseWriter, r *http.Request) (interface{}, error)

type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

type Wrapper struct {
	paramMappers  map[string]ParamMapper
	bodyExtractor RequestBodyExtractor
	bodyMapper    ResponseBodyMapper
	middlewares   []Middleware
	logger        log.Logger
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
		paramMappers:  mappers,
		bodyExtractor: bodyExtractor,
		bodyMapper:    bodyMapper,
		logger:        logger,
	}
}

func (m Wrapper) Endpoint(f interface{}) http.HandlerFunc {
	caller, err := NewCaller(f, m.bodyExtractor, m.bodyMapper, m.paramMappers)
	if err != nil {
		panic(err)
	}

	handler := caller.Handle
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r.Context(), w, r)
		if err != nil {
			m.logger.Error(r.Context(), err)
		}
	}
}

func (m Wrapper) WithMiddlewares(middlewares ...Middleware) Wrapper {
	return Wrapper{
		paramMappers:  m.paramMappers,
		bodyExtractor: m.bodyExtractor,
		bodyMapper:    m.bodyMapper,
		middlewares:   append(m.middlewares, middlewares...),
		logger:        m.logger,
	}

}
