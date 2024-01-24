package http

import (
	"context"
	"net/http"
)

type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type Middleware func(next HandlerFunc) HandlerFunc
