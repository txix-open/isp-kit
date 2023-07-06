package router

import (
	"context"
	"net/http"

	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/julienschmidt/httprouter"
)

type Params = httprouter.Params

type Router struct {
	router *httprouter.Router
}

func New() *Router {
	return &Router{
		router: httprouter.New(),
	}
}

func (r *Router) GET(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodGet, path, handler)
}

func (r *Router) POST(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodPost, path, handler)
}

func (r *Router) PUT(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodPut, path, handler)
}

func (r *Router) DELETE(path string, handler http.Handler) *Router {
	return r.Handler(http.MethodDelete, path, handler)
}

func (r *Router) Handler(method string, path string, handler http.Handler) *Router {
	r.router.Handler(method, path, r.withMetricEndpoint(path, handler))
	return r
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.router.ServeHTTP(writer, request)
}

func (r *Router) InternalRouter() *httprouter.Router {
	return r.router
}

func (r *Router) withMetricEndpoint(path string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := http_metrics.ServerEndpointToContext(request.Context(), path)
		request = request.WithContext(ctx)
		handler.ServeHTTP(writer, request)
	})
}

func ParamsFromRequest(http *http.Request) Params {
	return ParamsFromContext(http.Context())
}

func ParamsFromContext(ctx context.Context) Params {
	return httprouter.ParamsFromContext(ctx)
}
