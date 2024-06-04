package soap

import (
	"gitlab.txix.ru/isp/isp-kit/http"
	"gitlab.txix.ru/isp/isp-kit/http/endpoint"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/http_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/http/server_tracing"
	"gitlab.txix.ru/isp/isp-kit/validator"
)

const (
	defaultMaxRequestBodySize = 64 * 1024 * 1024
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...http.Middleware) endpoint.Wrapper {
	paramMappers := []endpoint.ParamMapper{
		endpoint.ContextParam(),
		endpoint.ResponseWriterParam(),
		endpoint.RequestParam(),
	}
	middlewares := append(
		[]http.Middleware{
			endpoint.MaxRequestBodySize(defaultMaxRequestBodySize),
			endpoint.RequestId(),
			endpoint.DefaultLog(logger),
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
