package cluster

import "fmt"

const (
	// ErrorConnection is the event name for connection errors.
	ErrorConnection = "ERROR_CONNECTION"
	// ConfigError is the event name for configuration errors.
	ConfigError = "ERROR_CONFIG"

	// ConfigSendConfigWhenConnected is the event for sending config when connected.
	ConfigSendConfigWhenConnected = "CONFIG:SEND_CONFIG_WHEN_CONNECTED"
	// ConfigSendConfigChanged is the event for config change notifications.
	ConfigSendConfigChanged = "CONFIG:SEND_CONFIG_CHANGED"

	// ConfigSendRoutesWhenConnected is the event for sending routes when connected.
	ConfigSendRoutesWhenConnected = "CONFIG:SEND_ROUTES_WHEN_CONNECTED"
	// ConfigSendRoutesChanged is the event for route change notifications.
	ConfigSendRoutesChanged = "CONFIG:SEND_ROUTES_CHANGED"

	// ModuleReady is the event sent when a module is ready.
	ModuleReady = "MODULE:READY"
	// ModuleSendRequirements is the event for sending module requirements.
	ModuleSendRequirements = "MODULE:SEND_REQUIREMENTS"
	// ModuleSendConfigSchema is the event for sending configuration schema.
	ModuleSendConfigSchema = "MODULE:SEND_CONFIG_SCHEMA"

	// ModuleConnectionSuffix is the suffix appended to module names for connection events.
	ModuleConnectionSuffix = "MODULE_CONNECTED"
)

// ModuleConnectedEvent generates the event name for a module connection event
// by combining the module name with the ModuleConnectionSuffix.
func ModuleConnectedEvent(moduleName string) string {
	return fmt.Sprintf("%s_%s", moduleName, ModuleConnectionSuffix)
}
