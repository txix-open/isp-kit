// Package schema provides JSON schema generation for Go structs with support for
// custom field properties, validation constraints, and tag-based configuration.
package schema

import (
	"github.com/txix-open/jsonschema"
)

// Schema is a pointer to a jsonschema.Schema, representing a JSON Schema document.
type Schema *jsonschema.Schema

// Generator creates JSON schemas from Go types with custom field reflection logic.
// It uses a jsonschema.Reflector configured with custom field name and property handlers.
type Generator struct {
	Reflector *jsonschema.Reflector
}

// NewGenerator creates a new Generator instance with default configuration.
// The reflector is set to expand structs and avoid using references ($ref).
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

// Generate creates a JSON schema from the provided Go value.
// The obj should typically be a pointer to a struct.
// Returns the generated Schema representing the JSON schema document.
func (g *Generator) Generate(obj any) Schema {
	return g.Reflector.Reflect(obj)
}
