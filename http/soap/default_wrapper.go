package soap

import (
	"github.com/integration-system/isp-kit/http/endpoint"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/integration-system/isp-kit/validator"
)

const (
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...endpoint.Middleware) endpoint.Wrapper {
	paramMappers := []endpoint.ParamMapper{
		endpoint.ContextParam(),
		endpoint.ResponseWriterParam(),
		endpoint.RequestParam(),
	}
	middlewares := append(
		[]endpoint.Middleware{
			endpoint.MaxRequestBodySize(defaultMaxRequestBodySize),
			endpoint.RequestId(),
			endpoint.DefaultLog(logger),
			endpoint.Metrics(http_metrics.NewServerStorage(metrics.DefaultRegistry)),
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
