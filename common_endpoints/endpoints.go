package common_endpoints

import (
	"fmt"
	"net/http"

	"github.com/txix-open/isp-kit/cluster"
)

type commonEndpointsCfg struct {
	swagger []byte
}

func CommonEndpoints(basePath string, opts ...CommonEndpointOption) []cluster.EndpointDescriptor {
	cfg := &commonEndpointsCfg{}
	for _, opt := range opts {
		opt(cfg)
	}

	endpoints := make([]cluster.EndpointDescriptor, 0)
	if len(cfg.swagger) > 0 {
		endpoints = append(endpoints, swaggerEndpoint(basePath, cfg.swagger))
	}

	return endpoints
}

func swaggerEndpoint(basePath string, swaggerYaml []byte) cluster.EndpointDescriptor {
	length := fmt.Sprintf("%d", len(swaggerYaml))
	return cluster.EndpointDescriptor{
		Path:             basePath + "/swagger",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodGet,
		Handler: func(res http.ResponseWriter) {
			res.Header().Set("Content-Type", "application/yaml")
			res.Header().Set("Content-Length", length)
			_, _ = res.Write(swaggerYaml)
		},
	}
}
