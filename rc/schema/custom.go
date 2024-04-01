package schema

import (
	"reflect"
	"sync"

	"github.com/txix-open/jsonschema"
)

type CustomGenerator func(field reflect.StructField, s *jsonschema.Schema)

type customSchema struct {
	mx        sync.RWMutex
	mapSchema map[string]CustomGenerator
}

var CustomGenerators = &customSchema{
	mapSchema: make(map[string]CustomGenerator),
}

func (c *customSchema) Register(name string, f CustomGenerator) {
	c.mx.Lock()
	c.mapSchema[name] = f
	c.mx.Unlock()
}

func (c *customSchema) Remove(name string) {
	c.mx.Lock()
	delete(c.mapSchema, name)
	c.mx.Unlock()
}

func (c *customSchema) getGeneratorByName(name string) CustomGenerator {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.mapSchema[name]
}
