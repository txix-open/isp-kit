// Package common_endpoints provides common HTTP endpoint descriptors for services.
// It offers a simple way to register standard endpoints like Swagger documentation.
package common_endpoints

// CommonEndpointOption is a functional option for configuring commonEndpointsCfg.
type CommonEndpointOption func(cfg *commonEndpointsCfg)

// WithSwaggerEndpoint configures the CommonEndpoints builder to include a Swagger endpoint.
// The swagger parameter should contain the JSON-formatted Swagger specification.
// The endpoint is exposed at {basePath}/swagger and is accessible via HTTP GET.
func WithSwaggerEndpoint(swagger []byte) CommonEndpointOption {
	return func(cfg *commonEndpointsCfg) {
		cfg.swagger = swagger
	}
}
