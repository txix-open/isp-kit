package cluster

// EndpointsResolver provides the list of endpoints for a module.
// Implement this interface when endpoints must be resolved dynamically at runtime.
// Use the Endpoints helper for static endpoint lists.
type EndpointsResolver interface {
	Endpoints() []EndpointDescriptor
}

// StaticEndpointResolver wraps a fixed slice of EndpointDescriptors.
// It implements EndpointsResolver for cases where the endpoint list is known at startup
// and does not change.
type StaticEndpointResolver []EndpointDescriptor

// Endpoints returns the underlying slice of EndpointDescriptors.
func (s StaticEndpointResolver) Endpoints() []EndpointDescriptor {
	return s
}

// Endpoints wraps a slice of EndpointDescriptors into a StaticEndpointResolver.
// Use this helper when constructing ModuleInfo or calling bootstrap.New with a known
// set of endpoints:
//
//	bootstrap.New(
//	    "1.0.0",
//	    &config{},
//	    cluster.Endpoints([]cluster.EndpointDescriptor{{Path: "/api", HttpMethod: "GET"}}),
//	    cluster.HttpTransport,
//	)
func Endpoints(arr []EndpointDescriptor) StaticEndpointResolver {
	return StaticEndpointResolver(arr)
}
