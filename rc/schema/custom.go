package schema

import (
	"reflect"
	"sync"

	"github.com/txix-open/jsonschema"
)

// CustomGenerator is a function type for customizing JSON schema generation.
// It receives the struct field and the schema to modify.
type CustomGenerator func(field reflect.StructField, s *jsonschema.Schema)

// customSchema manages the registry of custom schema generators.
// It is safe for concurrent use.
type customSchema struct {
	mx        sync.RWMutex
	mapSchema map[string]CustomGenerator
}

// CustomGenerators is the global registry for custom schema generators.
// Use Register to add custom generators and Remove to delete them.
var CustomGenerators = &customSchema{
	mapSchema: make(map[string]CustomGenerator),
}

// Register adds a custom schema generator to the registry.
// The name is used as the key to reference this generator from struct tags.
func (c *customSchema) Register(name string, f CustomGenerator) {
	c.mx.Lock()
	c.mapSchema[name] = f
	c.mx.Unlock()
}

// Remove deletes a custom schema generator from the registry by name.
func (c *customSchema) Remove(name string) {
	c.mx.Lock()
	delete(c.mapSchema, name)
	c.mx.Unlock()
}

// getGeneratorByName retrieves a custom generator from the registry by name.
// Returns nil if no generator is registered with the given name.
func (c *customSchema) getGeneratorByName(name string) CustomGenerator {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.mapSchema[name]
}
