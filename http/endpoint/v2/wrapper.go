package endpoint

import (
	"context"
	"io"
	"net/http"
	"slices"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/log"
)

type RequestBodyExtractor interface {
	Extract(ctx context.Context, reader io.Reader, ptr any) error
}

type ResponseBodyMapper interface {
	Map(ctx context.Context, result any, w http.ResponseWriter) error
}

type Wrapper struct {
	bodyExtractor RequestBodyExtractor
	bodyMapper    ResponseBodyMapper
	middlewares   []http2.Middleware
	logger        log.Logger
}

func NewWrapper(extractor RequestBodyExtractor, mapper ResponseBodyMapper, logger log.Logger) Wrapper {
	return Wrapper{
		bodyExtractor: extractor,
		bodyMapper:    mapper,
		middlewares:   nil,
		logger:        logger,
	}
}

func (m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper {
	m.middlewares = middlewares
	return m
}

func (m Wrapper) Endpoint(handler http2.HandlerFunc) http.HandlerFunc {
	for i := range slices.Backward(m.middlewares) {
		handler = m.middlewares[i](handler)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := handler(r.Context(), w, r)
		if err != nil {
			m.logger.Error(r.Context(), err)
		}
	}
}
