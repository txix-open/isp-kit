package endpoint

import (
	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/http/server_tracing"
	"github.com/txix-open/isp-kit/validator"
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
