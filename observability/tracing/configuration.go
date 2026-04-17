package tracing

// Config holds the configuration for the tracing provider.
type Config struct {
	// Enable determines whether tracing is enabled.
	Enable bool
	// Address specifies the OTLP collector endpoint address.
	Address string
	// ModuleName identifies the service name.
	ModuleName string
	// ModuleVersion specifies the service version.
	ModuleVersion string
	// Environment defines the deployment environment.
	Environment string
	// InstanceId uniquely identifies the service instance.
	InstanceId string
	// Attributes contains additional custom attributes for the resource.
	Attributes map[string]string
}
