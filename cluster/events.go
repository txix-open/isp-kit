package cluster

import (
	"fmt"
)

const (
	ErrorConnection = "ERROR_CONNECTION"
	ConfigError     = "ERROR_CONFIG"

	ConfigSendConfigWhenConnected = "CONFIG:SEND_CONFIG_WHEN_CONNECTED"
	ConfigSendConfigChanged       = "CONFIG:SEND_CONFIG_CHANGED"

	ConfigSendRoutesWhenConnected = "CONFIG:SEND_ROUTES_WHEN_CONNECTED"
	ConfigSendRoutesChanged       = "CONFIG:SEND_ROUTES_CHANGED"

	ModuleReady            = "MODULE:READY"
	ModuleSendRequirements = "MODULE:SEND_REQUIREMENTS"
	ModuleSendConfigSchema = "MODULE:SEND_CONFIG_SCHEMA"

	ModuleConnectionSuffix = "MODULE_CONNECTED"
)

func ModuleConnectedEvent(moduleName string) string {
	return fmt.Sprintf("%s_%s", moduleName, ModuleConnectionSuffix)
}
