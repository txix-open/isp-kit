package endpoint

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger, restMiddlewares ...Middleware) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	middlewares := append(
		[]Middleware{
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
