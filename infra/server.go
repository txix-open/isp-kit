package infra

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

type Server struct {
	mux *http.ServeMux
	s   *http.Server
}

func NewServer() *Server {
	mux := http.NewServeMux()
	return &Server{
		mux: mux,
		s:   &http.Server{Handler: mux},
	}
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.Handle(pattern, handler)
}

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

func (s *Server) Shutdown() error {
	return s.s.Shutdown(context.Background())
}
