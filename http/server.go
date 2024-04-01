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

type service struct {
	delegate atomic.Value
}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value, ok := s.delegate.Load().(http.Handler)
	if !ok {
		_ = apierrors.NewInternalServiceError(errors.New("handler is not initialized")).
			WriteError(w)
		return
	}

	value.ServeHTTP(w, r)
}

type ServerOption func(*Server)

func WithServer(server *http.Server) ServerOption {
	return func(srv *Server) {
		srv.server = server
	}
}

type Server struct {
	server  *http.Server
	service *service
}

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

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Upgrade(handler http.Handler) {
	s.service.delegate.Store(handler)
}

func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return errors.WithMessagef(err, "listen: %s", address)
	}

	return s.Serve(listener)
}

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
