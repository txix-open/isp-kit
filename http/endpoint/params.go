package endpoint

import (
	"context"
	"net/http"
)

// ContextParam creates a ParamMapper that provides context.Context to endpoint functions.
// This is automatically included in the default wrapper.
func ContextParam() ParamMapper {
	return ParamMapper{
		Type: "context.Context",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return ctx, nil
		},
	}
}

// ResponseWriterParam creates a ParamMapper that provides http.ResponseWriter to endpoint functions.
// This is automatically included in the default wrapper.
func ResponseWriterParam() ParamMapper {
	return ParamMapper{
		Type: "http.ResponseWriter",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return w, nil
		},
	}
}

// RequestParam creates a ParamMapper that provides *http.Request to endpoint functions.
// This is automatically included in the default wrapper.
func RequestParam() ParamMapper {
	return ParamMapper{
		Type: "*http.Request",
		Builder: func(ctx context.Context, w http.ResponseWriter, r *http.Request) (any, error) {
			return r, nil
		},
	}
}
