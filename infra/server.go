// Package infra provides a wrapper around the standard library's net/http
// HTTP server with graceful shutdown support.
package infra

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Server is a wrapper around http.Server that simplifies HTTP server setup
// and provides convenient methods for route registration and graceful shutdown.
type Server struct {
	mux *http.ServeMux
	s   *http.Server
}

// NewServer creates a new HTTP server instance.
func NewServer() *Server {
	mux := http.NewServeMux()
	return &Server{
		mux: mux,
		s:   &http.Server{Handler: mux, ReadHeaderTimeout: 1 * time.Second},
	}
}

// Handle registers an HTTP handler for the given pattern.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// HandleFunc registers an HTTP handler function for the given pattern.
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.Handle(pattern, handler)
}

// ListenAndServe starts the HTTP server on the specified address.
// It treats http.ErrServerClosed as a successful shutdown and returns nil.
// Returns an error wrapped with context if the server fails to start.
func (s *Server) ListenAndServe(address string) error {
	s.s.Addr = address
	err := s.s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return errors.WithMessagef(err, "listen and serve on %s", address)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server by closing all idle connections
// and waiting for outstanding requests to complete. It uses a background
// context with no timeout, so callers should implement their own timeout logic
// if needed.
func (s *Server) Shutdown() error {
	return s.s.Shutdown(context.Background())
}
