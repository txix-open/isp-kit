package schema

import (
	"reflect"
	"sync"

	"github.com/integration-system/jsonschema"
)

type Generator func(field reflect.StructField, t *jsonschema.Type)

type customSchema struct {
	mx        sync.RWMutex
	mapSchema map[string]Generator
}

var CustomGenerators = &customSchema{
	mapSchema: make(map[string]Generator),
}

func (c *customSchema) Register(name string, f Generator) {
	c.mx.Lock()
	c.mapSchema[name] = f
	c.mx.Unlock()
}

func (c *customSchema) Remove(name string) {
	c.mx.Lock()
	delete(c.mapSchema, name)
	c.mx.Unlock()
}

func (c *customSchema) getGeneratorByName(name string) Generator {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.mapSchema[name]
}
