package sentry

// Config holds the configuration for Sentry integration.
type Config struct {
	// Enable determines whether Sentry is enabled.
	Enable bool
	// Dsn is the Sentry Data Source Name. Required when Enable is true.
	Dsn string
	// ModuleName identifies the service or module sending events.
	ModuleName string
	// ModuleVersion is the version of the module.
	ModuleVersion string
	// Environment specifies the environment (e.g., "production", "development").
	Environment string
	// InstanceId uniquely identifies the running instance.
	InstanceId string
	// Tags are key-value pairs attached to all events.
	Tags map[string]string
}
