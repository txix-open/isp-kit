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

func init() {
	for tag, keyword := range boolKeywordTags {
		CustomGenerators.Register(tag, func(_ reflect.StructField, s *jsonschema.Schema) {
			if s.Extras == nil {
				s.Extras = make(map[string]any)
			}
			s.Extras[keyword] = true
		})
	}
}
