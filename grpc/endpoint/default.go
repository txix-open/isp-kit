package endpoint

import (
	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/grpc_metrics"
	"github.com/integration-system/isp-kit/observability/tracing/grpc/server_tracing"
	"github.com/integration-system/isp-kit/validator"
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
