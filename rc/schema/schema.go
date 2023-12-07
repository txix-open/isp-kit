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
	Reflector *jsonschema.Reflector
}

func NewGenerator() *Generator {
	return &Generator{
		Reflector: &jsonschema.Reflector{
			FieldNameReflector: GetNameAndRequiredFlag,
			FieldReflector:     SetProperties,
			ExpandedStruct:     true,
			DoNotReference:     true,
		},
	}
}

func (g *Generator) Generate(obj any) Schema {
	return g.Reflector.Reflect(obj)
}
