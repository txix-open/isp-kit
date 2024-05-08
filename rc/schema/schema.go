package schema

import (
	"github.com/txix-open/jsonschema"
)

type Schema *jsonschema.Schema

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
