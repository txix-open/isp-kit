package endpoint

import (
	"context"
	"net/http"
)

func ContextParam() ParamMapper {
	return ParamMapper{
		Type: "context.Context",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return ctx, nil
		},
	}
}

func ResponseWriterParam() ParamMapper {
	return ParamMapper{
		Type: "http.ResponseWriter",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return w, nil
		},
	}
}

func RequestParam() ParamMapper {
	return ParamMapper{
		Type: "*http.Request",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return r, nil
		},
	}
}
