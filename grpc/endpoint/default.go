package endpoint

import (
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/grpc_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/grpc/server_tracing"
	"github.com/txix-open/isp-kit/validator"
)

// DefaultWrapper creates a Wrapper with pre-configured middleware for observability.
// Includes request ID propagation, metrics collection, distributed tracing, error handling,
// and panic recovery. Uses JSON for request extraction and response mapping.
// Accepts additional middleware to be appended after the default ones.
func DefaultWrapper(logger log.Logger, restMiddlewares ...grpc.Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	metricStorage := grpc_metrics.NewServerStorage(metrics.DefaultRegistry)
	middlewares := append(
		[]grpc.Middleware{
			RequestId(),
			Metrics(metricStorage),
			server_tracing.NewConfig().Middleware(),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)
	return NewWrapper(
		paramMappers,
		JsonRequestExtractor{Validator: validator.Default},
		JsonResponseMapper{},
	).WithMiddlewares(middlewares...)
}
