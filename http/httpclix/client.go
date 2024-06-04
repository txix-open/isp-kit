package httpclix

import (
	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/http_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/http/client_tracing"
)

func Default(opts ...httpcli.Option) *httpcli.Client {
	opts = append([]httpcli.Option{
		httpcli.WithMiddlewares(
			RequestId(),
			Metrics(http_metrics.NewClientStorage(metrics.DefaultRegistry)),
			client_tracing.NewConfig().Middleware(),
		),
	}, opts...)
	return httpcli.New(opts...)
}
