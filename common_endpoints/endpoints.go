package common_endpoints

import (
	"net/http"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
)

// commonEndpointsCfg holds configuration for common endpoints.
type commonEndpointsCfg struct {
	swagger []byte
}

// CommonEndpoints creates a slice of cluster.EndpointDescriptor for common service endpoints.
// It accepts a basePath for URL path prefix and variadic CommonEndpointOption functions to configure endpoints.
// Returns a slice of endpoint descriptors that can be registered with a cluster router.
// Currently supports Swagger documentation endpoint when configured via WithSwaggerEndpoint.
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

// swaggerEndpoint creates an endpoint descriptor for serving Swagger documentation.
// The endpoint is registered at basePath+"/swagger" and returns the swagger JSON on HTTP GET.
// User authentication is not required for this endpoint.
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
