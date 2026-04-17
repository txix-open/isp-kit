// Package endpoint provides a higher-level abstraction for building gRPC handlers.
// It enables reflection-based function wrapping with automatic request/response
// marshaling, parameter injection, and middleware support.
//
// The package integrates with the ISP kit framework's logging, metrics, and
// validation systems for comprehensive observability and error handling.
package endpoint

import (
	"context"
	"reflect"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
)

// GrpcError is an interface for errors that can be converted to gRPC status errors.
// Implementations should return a gRPC status error with appropriate details.
type GrpcError interface {
	GrpcStatusError() error
}

// RequestBodyExtractor defines the interface for extracting request bodies from gRPC messages.
// Implementations should unmarshal the message body into the target type and return a reflect.Value.
type RequestBodyExtractor interface {
	// Extract unmarshals the message body into an instance of reqBodyType.
	// Returns the reflect.Value of the created instance and any unmarshaling or validation errors.
	Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error)
}

// ResponseBodyMapper defines the interface for mapping handler results to gRPC messages.
// Implementations should marshal the result into a gRPC message body.
type ResponseBodyMapper interface {
	// Map converts a handler result into a gRPC message.
	// Returns the message and any marshaling errors.
	Map(result any) (*isp.Message, error)
}

// ParamBuilder defines a function that builds a parameter value from context and message.
// Used by ParamMapper to inject dependencies like context, auth data, etc.
type ParamBuilder func(ctx context.Context, message *isp.Message) (any, error)

// ParamMapper maps a parameter type string to a builder function.
// Used to inject dependencies into handler functions automatically.
type ParamMapper struct {
	// Type is the string representation of the parameter type.
	Type string
	// Builder is the function that creates the parameter value.
	Builder ParamBuilder
}

// Wrapper is a high-level handler factory that wraps functions with middleware and
// automatic request/response handling. It uses reflection to analyze function signatures
// and inject dependencies like context and auth data.
type Wrapper struct {
	ParamMappers  map[string]ParamMapper
	BodyExtractor RequestBodyExtractor
	BodyMapper    ResponseBodyMapper
	Middlewares   []grpc.Middleware
}

// NewWrapper creates a new Wrapper with the specified configuration.
// Accepts variadic paramMappers, a body extractor, and a body mapper.
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

// Endpoint wraps a function to create a gRPC HandlerFunc.
// The function is analyzed via reflection to determine parameter types and return values.
// Middleware is applied in reverse order (last added, first executed).
// Panics if the function signature is invalid or cannot be wrapped.
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

// WithMiddlewares returns a new Wrapper with additional middleware appended.
// The original Wrapper is not modified.
func (m Wrapper) WithMiddlewares(middlewares ...grpc.Middleware) Wrapper {
	return Wrapper{
		ParamMappers:  m.ParamMappers,
		BodyExtractor: m.BodyExtractor,
		BodyMapper:    m.BodyMapper,
		Middlewares:   append(m.Middlewares, middlewares...),
	}
}
