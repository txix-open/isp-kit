package common_endpoints

import (
	"encoding/base64"
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

func swaggerEndpoint(basePath string, swagger []byte) cluster.EndpointDescriptor {
	encodedSwagger := base64.StdEncoding.EncodeToString(swagger)

	return cluster.EndpointDescriptor{
		Path:             basePath + "/swagger",
		Inner:            false,
		UserAuthRequired: false,
		HttpMethod:       http.MethodGet,
		Handler: func() string {
			return encodedSwagger
		},
	}
}
