package infra

import (
	"context"
	"net/http"
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

func (s *Server) ListenAndServe(address string) error {
	s.s.Addr = address
	return s.s.ListenAndServe()
}

func (s *Server) Shutdown() error {
	return s.s.Shutdown(context.Background())
}
