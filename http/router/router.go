// Package router provides HTTP request routing with metrics integration.
// It wraps the julienschmidt/httprouter library and adds automatic endpoint metrics collection.
package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
)

// Params is an alias for httprouter.Params, representing URL path parameters.
type Params = httprouter.Params

// Router wraps httprouter.Router with automatic metrics collection for each endpoint.
// It supports method-based routing (GET, POST, PUT, DELETE) and fluent API for route configuration.
type Router struct {
	router *httprouter.Router
}

// New creates a new Router instance with a fresh httprouter backend.
func New() *Router {
	return &Router{
		router: httprouter.New(),
	}
}

// GET registers an HTTP GET handler for the specified path.
// Returns the Router for fluent chaining.
func (r *Router) GET(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodGet, path, handler)
}

// POST registers an HTTP POST handler for the specified path.
// Returns the Router for fluent chaining.
func (r *Router) POST(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodPost, path, handler)
}

// PUT registers an HTTP PUT handler for the specified path.
// Returns the Router for fluent chaining.
func (r *Router) PUT(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodPut, path, handler)
}

// DELETE registers an HTTP DELETE handler for the specified path.
// Returns the Router for fluent chaining.
func (r *Router) DELETE(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodDelete, path, handler)
}

// Handler registers a handler for the specified HTTP method and path.
// It automatically wraps the handler with metrics collection.
// Returns the Router for fluent chaining.
func (r *Router) Handler(method string, path string, handler http.Handler) *Router {
	r.router.Handler(method, path, r.withMetricEndpoint(method, path, handler))
	return r
}

// ServeHTTP implements the http.Handler interface and delegates to the underlying httprouter.
func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.router.ServeHTTP(writer, request)
}

// InternalRouter returns the underlying httprouter instance for advanced configuration.
func (r *Router) InternalRouter() *httprouter.Router {
	return r.router
}

// withMetricEndpoint wraps a handler to add endpoint metrics to the request context.
// The endpoint is formatted as "METHOD /path" for metrics identification.
func (r *Router) withMetricEndpoint(method string, path string, handler http.Handler) http.Handler {
	endpoint := fmt.Sprintf("%s %s", method, path)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := http_metrics.ServerEndpointToContext(request.Context(), endpoint)
		request = request.WithContext(ctx)
		handler.ServeHTTP(writer, request)
	})
}

// ParamsFromRequest extracts URL path parameters from an http.Request.
// It retrieves the parameters from the request context.
func ParamsFromRequest(http *http.Request) Params {
	return ParamsFromContext(http.Context())
}

// ParamsFromContext extracts URL path parameters from a context.
// It returns empty params if none are present in the context.
func ParamsFromContext(ctx context.Context) Params {
	return httprouter.ParamsFromContext(ctx)
}
