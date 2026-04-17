package cluster

import (
	"encoding/json"
)

const (
	RequiredAdminPermission = "reqAdminPerm"

	GrpcTransport  = "grpc"
	HttpTransport  = "http"
	EmptyTransport = "empty"
)

// AddressConfiguration represents the IP address and port for a service endpoint.
type AddressConfiguration struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// ConfigData contains configuration data for a module, including version information
// and JSON-encoded schema and configuration payloads.
type ConfigData struct {
	Version string
	Schema  json.RawMessage
	Config  json.RawMessage
}

// ModuleInfo describes a module's metadata, including its name, version, transport
// configuration, exposed endpoints, and metrics autodiscovery settings.
type ModuleInfo struct {
	ModuleName           string
	ModuleVersion        string
	LibVersion           string
	Transport            string
	GrpcOuterAddress     AddressConfiguration
	Endpoints            []EndpointDescriptor
	MetricsAutodiscovery *MetricsAutodiscovery
}

// RoutingConfig represents a list of backend declarations for routing configuration.
type RoutingConfig []BackendDeclaration

// BackendDeclaration describes a backend service, including its metadata, endpoints,
// dependencies, and network configuration.
type BackendDeclaration struct {
	ModuleName           string
	Version              string
	LibVersion           string
	Transport            string
	Endpoints            []EndpointDescriptor
	RequiredModules      []ModuleDependency
	Address              AddressConfiguration
	MetricsAutodiscovery *MetricsAutodiscovery
}

// EndpointDescriptor describes an HTTP endpoint, including its path, method, authentication
// requirements, and optional metadata. The Handler field is excluded from JSON serialization.
type EndpointDescriptor struct {
	Path             string
	Inner            bool
	UserAuthRequired bool
	HttpMethod       string
	Extra            map[string]any
	Handler          any `json:"-"`
}

// ModuleRequirements specifies the modules and routes that a module depends on.
type ModuleRequirements struct {
	RequiredModules []string
	RequireRoutes   bool
}

// ModuleDependency represents a dependency on another module, indicating whether
// the dependency is required for operation.
type ModuleDependency struct {
	Name     string
	Required bool
}

// MetricsAutodiscovery configuration for automatic metrics endpoint discovery,
// including the address and optional Prometheus labels.
type MetricsAutodiscovery struct {
	Address string
	Labels  map[string]string
}

// RequireAdminPermission creates endpoint metadata that specifies an admin permission
// requirement for the endpoint. The permission string is stored under the RequiredAdminPermission key.
func RequireAdminPermission(perm string) map[string]any {
	return map[string]any{
		RequiredAdminPermission: perm,
	}
}

// GetRequiredAdminPermission extracts the admin permission requirement from an
// EndpointDescriptor. Returns the permission string and a boolean indicating whether
// the permission requirement was found.
func GetRequiredAdminPermission(desc EndpointDescriptor) (string, bool) {
	if len(desc.Extra) == 0 {
		return "", false
	}
	value, ok := desc.Extra[RequiredAdminPermission].(string)
	return value, ok
}
