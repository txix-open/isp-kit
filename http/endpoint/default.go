package endpoint

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		ResponseWriterParam(),
		RequestParam(),
	}
	middlewares := append(
		[]Middleware{
			RequestId(),
			RequestInfo(),
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
