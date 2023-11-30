package schema

import (
	"github.com/integration-system/jsonschema"
)

type Schema *jsonschema.Schema

type ConfigSchema struct {
	Version       string         `json:"version"`
	Schema        Schema         `json:"schema"`
	DefaultConfig map[string]any `json:"config"`
}

func GenerateConfigSchema(cfgPtr any) Schema {
	ref := jsonschema.Reflector{
		FieldNameReflector: GetNameAndRequiredFlag,
		FieldReflector:     SetProperties,
		ExpandedStruct:     true,
	}
	s := ref.Reflect(cfgPtr)
	s.Title = "Remote config"
	s.Version = ""
	return s
}
