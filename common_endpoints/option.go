package common_endpoints

type CommonEndpointOption func(cfg *commonEndpointsCfg)

func WithSwaggerEndpoint(swagger []byte) CommonEndpointOption {
	return func(cfg *commonEndpointsCfg) {
		cfg.swagger = swagger
	}
}
