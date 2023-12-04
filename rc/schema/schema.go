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

type Generator struct {
	ref *jsonschema.Reflector
}

func NewGenerator(ref *jsonschema.Reflector) *Generator {
	return &Generator{
		ref: ref,
	}
}

func (g *Generator) Generate(obj any) Schema {
	return g.ref.Reflect(obj)
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
