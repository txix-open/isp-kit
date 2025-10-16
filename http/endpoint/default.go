package endpoint

import (
	"context"
	"io"

	"github.com/txix-open/isp-kit/http"
	v2 "github.com/txix-open/isp-kit/http/endpoint/v2"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/http/server_tracing"
	"github.com/txix-open/isp-kit/validator"
)

func DefaultWrapper(logger log.Logger, logMiddleware LogMiddleware, restMiddlewares ...http.Middleware) v2.Wrapper {
	middlewares := append(
		[]http.Middleware{
			MaxRequestBodySize(defaultMaxRequestBodySize),
			RequestId(),
			http.Middleware(logMiddleware),
			Metrics(http_metrics.NewServerStorage(metrics.DefaultRegistry)),
			server_tracing.NewConfig().Middleware(),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)

	return v2.NewWrapper(
		extractorWrapper{inner: JsonRequestExtractor{Validator: validator.Default}},
		JsonResponseMapper{},
		logger,
	).WithMiddlewares(middlewares...)
}

type extractorWrapper struct {
	inner JsonRequestExtractor
}

func (e extractorWrapper) Extract(_ context.Context, reader io.Reader, ptr any) error {
	err := e.inner.extract(reader, ptr)
	if err != nil {
		return err
	}
	return e.inner.validate(ptr)
}
