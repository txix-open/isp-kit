package endpoint

import (
	"gitlab.txix.ru/isp/isp-kit/http"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/http_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/http/server_tracing"
	"gitlab.txix.ru/isp/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...http.Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		ResponseWriterParam(),
		RequestParam(),
	}
	middlewares := append(
		[]http.Middleware{
			MaxRequestBodySize(defaultMaxRequestBodySize),
			RequestId(),
			DefaultLog(logger),
			Metrics(http_metrics.NewServerStorage(metrics.DefaultRegistry)),
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
		logger,
	).WithMiddlewares(middlewares...)
}
