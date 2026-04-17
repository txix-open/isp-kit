// Package soap provides SOAP message handling for XML-based web services.
// It implements the SOAP 1.1 protocol with support for envelopes, headers, bodies, and faults.
package soap

import (
	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/http/server_tracing"
	"github.com/txix-open/isp-kit/validator"
)

const (
	// defaultMaxRequestBodySize is the default maximum allowed SOAP request body size (64MB).
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

// DefaultWrapper creates a pre-configured endpoint.Wrapper for SOAP services.
// It includes request logging, metrics collection, tracing, error handling, and recovery.
// The default maximum request body size is 64MB.
func DefaultWrapper(logger log.Logger, logMiddleware endpoint.LogMiddleware, restMiddlewares ...http.Middleware) endpoint.Wrapper {
	paramMappers := []endpoint.ParamMapper{
		endpoint.ContextParam(),
		endpoint.ResponseWriterParam(),
		endpoint.RequestParam(),
	}
	middlewares := append(
		[]http.Middleware{
			endpoint.MaxRequestBodySize(defaultMaxRequestBodySize),
			endpoint.RequestId(),
			http.Middleware(logMiddleware),
			endpoint.Metrics(http_metrics.NewServerStorage(metrics.DefaultRegistry)),
			server_tracing.NewConfig().Middleware(),
			ErrorHandler(logger),
			endpoint.Recovery(),
		},
		restMiddlewares...,
	)

	return endpoint.NewWrapper(
		paramMappers,
		RequestExtractor{Validator: validator.Default},
		ResponseMapper{},
		logger,
	).WithMiddlewares(middlewares...)
}
