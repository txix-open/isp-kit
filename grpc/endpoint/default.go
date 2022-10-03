package endpoint

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/grpc_metrics"
	"github.com/integration-system/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	metricStorage := grpc_metrics.NewServerStorage(metrics.DefaultRegistry)
	middlewares := append(
		[]Middleware{
			Metrics(metricStorage),
			RequestId(),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)
	return NewWrapper(
		paramMappers,
		JsonRequestExtractor{validator: validator.Default},
		JsonResponseMapper{},
	).WithMiddlewares(middlewares...)
}
