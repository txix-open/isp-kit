package common_endpoints

import (
	"net/http"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
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

func swaggerEndpoint(basePath string, swaggerJson json.RawMessage) cluster.EndpointDescriptor {
	return cluster.EndpointDescriptor{
		Path:             basePath + "/swagger",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodGet,
		Handler: func() json.RawMessage {
			return swaggerJson
		},
	}
}
