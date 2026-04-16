package http

import (
	"context"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/apierrors"
	"github.com/txix-open/isp-kit/log"
)

// service is an internal wrapper that delegates to the actual HTTP handler.
// It uses atomic.Value for thread-safe handler replacement.
type service struct {
	delegate atomic.Value
}

// ServeHTTP implements the http.Handler interface and delegates to the stored handler.
// It returns an internal service error if the handler is not initialized.
func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value, ok := s.delegate.Load().(http.Handler)
	if !ok {
		_ = apierrors.NewInternalServiceError(errors.New("handler is not initialized")).
			WriteError(w)
		return
	}

	value.ServeHTTP(w, r)
}

// ServerOption is a function that configures a Server instance.
type ServerOption func(*Server)

// WithServer allows providing a custom *http.Server for advanced configuration.
func WithServer(server *http.Server) ServerOption {
	return func(srv *Server) {
		srv.server = server
	}
}

// Server wraps an http.Server with additional functionality for handler management.
// It provides graceful shutdown and handler upgrade capabilities.
type Server struct {
	server  *http.Server
	service *service
}

// NewServer creates a new HTTP server with the specified logger and options.
// Default timeouts are set: 3 seconds for ReadHeaderTimeout and 120 seconds for IdleTimeout.
// Server is safe for concurrent use.
func NewServer(logger log.Logger, opts ...ServerOption) *Server {
	s := &Server{
		server: &http.Server{
			ReadHeaderTimeout: 3 * time.Second,
			IdleTimeout:       120 * time.Second,
			ErrorLog: log.StdLoggerWithLevel(
				logger,
				log.WarnLevel,
				log.String("worker", "http server"),
			),
		},
		service: &service{
			delegate: atomic.Value{},
		},
	}

	for _, opts := range opts {
		opts(s)
	}

	s.server.Handler = s.service

	return s
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
// It delegates to the underlying http.Server.Shutdown method.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// Upgrade replaces the current HTTP handler with a new one.
// This method is safe for concurrent use.
func (s *Server) Upgrade(handler http.Handler) {
	s.service.delegate.Store(handler)
}

// ListenAndServe starts the server on the specified address.
// It creates a TCP listener and delegates to Serve.
func (s *Server) ListenAndServe(address string) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", address)
	if err != nil {
		return errors.WithMessagef(err, "listen: %s", address)
	}

	return s.Serve(listener)
}

// Serve accepts incoming connections on the specified listener and handles requests.
// It returns nil when the server is gracefully shut down.
func (s *Server) Serve(listener net.Listener) error {
	err := s.server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return errors.WithMessage(err, "serve http")
	}
	return nil
}
