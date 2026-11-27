package schema

import (
	"reflect"

	"github.com/txix-open/jsonschema"
)

const (
	UIExposedTag     = "uiExposed"
	UIExposedKeyword = "x-ui-exposed"

	SensitiveTag     = "sensitive"
	SensitiveKeyword = "x-sensitive"
)

var boolKeywordTags = map[string]string{
	UIExposedTag: UIExposedKeyword,
	SensitiveTag: SensitiveKeyword,
}

// registerBoolKeywordGenerators registers the built-in bool-keyword
// generators (uiExposed, sensitive) into c.
func registerBoolKeywordGenerators(c *customSchema) {
	for tag, keyword := range boolKeywordTags {
		c.Register(tag, func(_ reflect.StructField, s *jsonschema.Schema) {
			if s.Extras == nil {
				s.Extras = make(map[string]any)
			}
			s.Extras[keyword] = true
		})
	}
}
