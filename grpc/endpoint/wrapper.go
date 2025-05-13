package endpoint

import (
	"context"
	"reflect"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
)

type GrpcError interface {
	GrpcStatusError() error
}

type RequestBodyExtractor interface {
	Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error)
}

type ResponseBodyMapper interface {
	Map(result any) (*isp.Message, error)
}

type ParamBuilder func(ctx context.Context, message *isp.Message) (any, error)

type ParamMapper struct {
	Type    string
	Builder ParamBuilder
}

type Wrapper struct {
	ParamMappers  map[string]ParamMapper
	BodyExtractor RequestBodyExtractor
	BodyMapper    ResponseBodyMapper
	Middlewares   []grpc.Middleware
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
		ParamMappers:  mappers,
		BodyExtractor: bodyExtractor,
		BodyMapper:    bodyMapper,
	}
}

func (m Wrapper) Endpoint(f any) grpc.HandlerFunc {
	caller, err := NewCaller(f, m.BodyExtractor, m.BodyMapper, m.ParamMappers)
	if err != nil {
		panic(err)
	}

	handler := caller.Handle
	for i := len(m.Middlewares) - 1; i >= 0; i-- {
		handler = m.Middlewares[i](handler)
	}
	return handler
}

func (m Wrapper) WithMiddlewares(middlewares ...grpc.Middleware) Wrapper {
	return Wrapper{
		ParamMappers:  m.ParamMappers,
		BodyExtractor: m.BodyExtractor,
		BodyMapper:    m.BodyMapper,
		Middlewares:   append(m.Middlewares, middlewares...),
	}
}
