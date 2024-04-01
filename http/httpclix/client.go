package httpclix

import (
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/http/client_tracing"
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
