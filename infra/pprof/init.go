package pprof

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

// go tool pprof -http=0.0.0.0:6061 http://localhost:10000/internal/debug/pprof/allocs
// go tool pprof -http=0.0.0.0:6061 http://localhost:10000/internal/debug/pprof/profile?seconds=10
// go tool pprof http://localhost:10000/internal/debug/pprof/profile?seconds=10 > profile.out

type Muxer interface {
	Handle(patter string, handler http.Handler)
}

func RegisterHandlers(prefix string, muxer Muxer) {
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/", prefix), cutPrefix(prefix, pprof.Index))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/cmdline", prefix), cutPrefix(prefix, pprof.Cmdline))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/profile", prefix), cutPrefix(prefix, pprof.Profile))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/symbol", prefix), cutPrefix(prefix, pprof.Symbol))
	muxer.Handle(fmt.Sprintf("%s/debug/pprof/trace", prefix), cutPrefix(prefix, pprof.Trace))
}

func Endpoints(prefix string) []string {
	return []string{
		fmt.Sprintf("%s/debug/pprof/", prefix),
		fmt.Sprintf("%s/debug/pprof/cmdline", prefix),
		fmt.Sprintf("%s/debug/pprof/profile", prefix),
		fmt.Sprintf("%s/debug/pprof/symbol", prefix),
		fmt.Sprintf("%s/debug/pprof/trace", prefix),
	}
}

func cutPrefix(prefix string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		request.URL.Path = request.URL.Path[len(prefix):]
		handler.ServeHTTP(writer, request)
	})
}
