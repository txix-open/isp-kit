package endpoint

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/integration-system/isp-kit/validator"
)

type Logger interface {
	log.Logger
	Enabled(level log.Level) bool
}

func DefaultWrapper(logger Logger, restMiddlewares ...Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		ResponseWriterParam(),
		RequestParam(),
	}
	middlewares := append(
		[]Middleware{
			MaxRequestBodySize(defaultMaxRequestBodySize),
			RequestId(),
			RequestInfo(),
			DefaultLog(logger),
			Metrics(http_metrics.NewStorage(metrics.DefaultRegistry)),
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
