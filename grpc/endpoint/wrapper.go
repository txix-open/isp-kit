package endpoint

import (
	"context"
	"reflect"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
)

type Middleware func(next grpc.HandlerFunc) grpc.HandlerFunc

type RequestBodyExtractor interface {
	Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error)
}

type ResponseBodyMapper interface {
	Map(result interface{}) (*isp.Message, error)
}

type ParamBuilder func(ctx context.Context, message *isp.Message) (interface{}, error)

type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

type Wrapper struct {
	paramMappers  map[string]ParamMapper
	bodyExtractor RequestBodyExtractor
	bodyMapper    ResponseBodyMapper
	middlewares   []Middleware
}

func NewWrapper(
	paramMappers []ParamMapper,
	bodyExtractor RequestBodyExtractor,
	bodyMapper ResponseBodyMapper,
) Wrapper {
	mappers := make(map[string]ParamMapper)
	for _, mapper := range paramMappers {
		mappers[mapper.Type] = mapper
	}
	return Wrapper{
		paramMappers:  mappers,
		bodyExtractor: bodyExtractor,
		bodyMapper:    bodyMapper,
	}
}

func (m Wrapper) Endpoint(f interface{}) grpc.HandlerFunc {
	caller, err := NewCaller(f, m.bodyExtractor, m.bodyMapper, m.paramMappers)
	if err != nil {
		panic(err)
	}

	handler := caller.Handle
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		handler = m.middlewares[i](handler)
	}
	return handler
}

func (m Wrapper) WithMiddlewares(middlewares ...Middleware) Wrapper {
	return Wrapper{
		paramMappers:  m.paramMappers,
		bodyExtractor: m.bodyExtractor,
		bodyMapper:    m.bodyMapper,
		middlewares:   append(m.middlewares, middlewares...),
	}

}
