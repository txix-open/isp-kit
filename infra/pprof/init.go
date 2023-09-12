package pprof

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

type Muxer interface {
	HandleFunc(patter string, handler http.HandlerFunc)
}

func RegisterHandlers(prefix string, muxer Muxer) {
	muxer.HandleFunc(fmt.Sprintf("%s/debug/pprof/", prefix), pprof.Index)
	muxer.HandleFunc(fmt.Sprintf("%s/debug/pprof/cmdline", prefix), pprof.Cmdline)
	muxer.HandleFunc(fmt.Sprintf("%s/debug/pprof/profile", prefix), pprof.Profile)
	muxer.HandleFunc(fmt.Sprintf("%s/debug/pprof/symbol", prefix), pprof.Symbol)
	muxer.HandleFunc(fmt.Sprintf("%s/debug/pprof/trace", prefix), pprof.Trace)
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
