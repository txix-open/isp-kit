package endpoint

import (
	"github.com/integration-system/isp-kit/log"
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
			RequestId(),
			RequestInfo(),
			DefaultLog(logger),
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
