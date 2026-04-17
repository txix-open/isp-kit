package rc

import (
	"github.com/txix-open/isp-kit/rc/schema"
)

// GenerateConfigSchema creates a JSON schema from a configuration struct pointer.
// It sets the title to "Remote config" and returns the generated schema.
// The cfgPtr should be a pointer to a struct that represents the configuration.
func GenerateConfigSchema(cfgPtr any) schema.Schema {
	s := schema.NewGenerator().Generate(cfgPtr)
	s.Title = "Remote config"
	s.Version = ""
	return s
}
