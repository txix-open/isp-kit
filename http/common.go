// Package http provides core HTTP server functionality and types for building REST and SOAP services.
// It includes a server implementation, handler wrapper, and middleware support for request processing.
package http

import (
	"context"
	"net/http"
)

// HandlerFunc is the signature for HTTP handlers that process requests in a context-aware manner.
// Handlers should return an error to indicate processing failure.
type HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Middleware is a function that wraps a HandlerFunc to add cross-cutting concerns
// such as logging, authentication, or metrics collection.
type Middleware func(next HandlerFunc) HandlerFunc
