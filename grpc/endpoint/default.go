package endpoint

import (
	"gitlab.txix.ru/isp/isp-kit/grpc"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/grpc_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/grpc/server_tracing"
	"gitlab.txix.ru/isp/isp-kit/validator"
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
