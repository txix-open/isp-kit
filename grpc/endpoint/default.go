package endpoint

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger) Wrapper {
	paramMappers := []ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	return NewWrapper(
		paramMappers,
		JsonRequestExtractor{validator: validator.Default},
		JsonResponseMapper{},
	).WithMiddlewares(
		RequestId(),
		ErrorHandler(logger),
		Recovery(),
	)
}
