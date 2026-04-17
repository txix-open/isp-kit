// Package pprof provides utilities for integrating Go's pprof profiling
// tools with an HTTP server. It registers the standard pprof endpoints
// under a configurable URL prefix.
package pprof

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

// Muxer defines an interface for HTTP multiplexers that can register handlers.
// This allows the package to work with various HTTP routing implementations.
type Muxer interface {
	Handle(pattern string, handler http.Handler)
}

// RegisterHandlers registers pprof handlers with the provided muxer under
// the specified URL prefix. This enables access to profiling endpoints such
// as /debug/pprof/, /debug/pprof/profile, and others.
//
// Example:
//
//	pprof.RegisterHandlers("/api", srv)
func RegisterHandlers(prefix string, muxer Muxer) {
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/", prefix), cutPrefix(prefix, pprof.Index))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/cmdline", prefix), cutPrefix(prefix, pprof.Cmdline))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/profile", prefix), cutPrefix(prefix, pprof.Profile))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/symbol", prefix), cutPrefix(prefix, pprof.Symbol))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/trace", prefix), cutPrefix(prefix, pprof.Trace))
}

// Endpoints returns a list of all pprof endpoint paths for the given prefix.
// This is useful for documentation or verification purposes.
func Endpoints(prefix string) []string {
	return []string{
		fmt.Sprintf("%s/debug/pprof/", prefix),
		fmt.Sprintf("%s/debug/pprof/cmdline", prefix),
		fmt.Sprintf("%s/debug/pprof/profile", prefix),
		fmt.Sprintf("%s/debug/pprof/symbol", prefix),
		fmt.Sprintf("%s/debug/pprof/trace", prefix),
	}
}

// cutPrefix creates a middleware handler that strips the specified prefix
// from the request URL path before passing it to the underlying handler.
// This allows pprof handlers to work correctly when mounted under a custom prefix.
func cutPrefix(prefix string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.URL.Path = request.URL.Path[len(prefix):]
		handler.ServeHTTP(writer, request)
	})
}
