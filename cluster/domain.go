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

type AddressConfiguration struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type ConfigData struct {
	Version string
	Schema  json.RawMessage
	Config  json.RawMessage
}

type ModuleInfo struct {
	ModuleName           string
	ModuleVersion        string
	LibVersion           string
	Transport            string
	GrpcOuterAddress     AddressConfiguration
	Endpoints            []EndpointDescriptor
	MetricsAutodiscovery *MetricsAutodiscovery
}

type RoutingConfig []BackendDeclaration

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

type EndpointDescriptor struct {
	Path             string
	Inner            bool
	UserAuthRequired bool
	HttpMethod       string
	Extra            map[string]any
	Handler          any `json:"-"`
}

type ModuleRequirements struct {
	RequiredModules []string
	RequireRoutes   bool
}

type ModuleDependency struct {
	Name     string
	Required bool
}

type MetricsAutodiscovery struct {
	Address string
	Labels  map[string]string
}

func RequireAdminPermission(perm string) map[string]any {
	return map[string]any{
		RequiredAdminPermission: perm,
	}
}

func GetRequiredAdminPermission(desc EndpointDescriptor) (string, bool) {
	if len(desc.Extra) == 0 {
		return "", false
	}
	value, ok := desc.Extra[RequiredAdminPermission].(string)
	return value, ok
}
