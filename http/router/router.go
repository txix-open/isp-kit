package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
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
	r.router.Handler(method, path, r.withMetricEndpoint(method, path, handler))
	return r
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.router.ServeHTTP(writer, request)
}

func (r *Router) InternalRouter() *httprouter.Router {
	return r.router
}

func (r *Router) withMetricEndpoint(method string, path string, handler http.Handler) http.Handler {
	endpoint := fmt.Sprintf("%s %s", method, path)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := http_metrics.ServerEndpointToContext(request.Context(), endpoint)
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
