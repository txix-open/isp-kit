package endpoint

import (
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/grpc_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/grpc/server_tracing"
	"github.com/txix-open/isp-kit/validator"
)

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
