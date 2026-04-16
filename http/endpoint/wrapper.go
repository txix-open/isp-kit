// Package endpoint provides a high-level abstraction for building HTTP endpoints
// with automatic request/response handling, validation, and middleware support.
// It uses Go generics to create type-safe endpoint implementations.
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

// HttpError is an interface for errors that can write themselves to an HTTP response.
// Implementations should handle their own HTTP status codes and response formatting.
type HttpError interface {
	WriteError(w http.ResponseWriter) error
}

// RequestBodyExtractor is responsible for extracting and unmarshaling request bodies.
// It supports both reflection-based and direct pointer-based extraction.
type RequestBodyExtractor interface {
	// Extract extracts the request body into a reflect.Value of the specified type.
	Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error)
	// ExtractV2 extracts the request body into a pointer directly.
	ExtractV2(ctx context.Context, reader io.Reader, ptr any) error
}

// ResponseBodyMapper is responsible for marshaling and writing response bodies.
type ResponseBodyMapper interface {
	// Map marshals the result and writes it to the http.ResponseWriter.
	Map(ctx context.Context, result any, w http.ResponseWriter) error
}

// ParamBuilder is a function that builds a parameter value from the request context.
type ParamBuilder func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error)

// ParamMapper maps a parameter type to its builder function.
type ParamMapper struct {
	// Type is the string representation of the parameter type.
	Type string
	// Builder creates the parameter value from the request context.
	Builder ParamBuilder
}

// Wrapper is the core component for building HTTP endpoints.
// It combines parameter mapping, body extraction, response mapping, and middleware support.
type Wrapper struct {
	ParamMappers  map[string]ParamMapper
	BodyExtractor RequestBodyExtractor
	BodyMapper    ResponseBodyMapper
	Middlewares   []http2.Middleware
	Logger        log.Logger
}

// NewWrapper creates a new Wrapper with the specified configuration.
// Parameter mappers are stored in a map for efficient type lookup.
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

// Endpoint wraps a function as an http.HandlerFunc.
//
// Use EndpointV2 with New, NewWithoutResponse, NewWithRequest, or NewDefaultHttp instead.
// EndpointV2 avoids reflection overhead by using the Wrappable interface.
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

// EndpointV2 wraps a Wrappable implementation as an http.HandlerFunc.
// It applies middleware in reverse order (last middleware executes first).
// This version avoids reflection overhead and is preferred for production use.
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

// WithMiddlewares returns a new Wrapper with additional middleware appended.
// The original Wrapper remains unchanged (immutable).
func (m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper {
	return Wrapper{
		ParamMappers:  m.ParamMappers,
		BodyExtractor: m.BodyExtractor,
		BodyMapper:    m.BodyMapper,
		Middlewares:   append(m.Middlewares, middlewares...),
		Logger:        m.Logger,
	}
}
