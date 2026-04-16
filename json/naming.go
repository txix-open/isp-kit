package json

import (
	"strings"
	"unicode"

	"github.com/json-iterator/go"
)

// namingStrategyExtension is a jsoniter extension that converts exported struct field
// names from PascalCase to camelCase during JSON serialization/deserialization.
//
// It skips fields that:
//   - Start with a lowercase letter or underscore
//   - Have an explicit "json" tag (non-empty or "-")
type namingStrategyExtension struct {
	jsoniter.DummyExtension

	translate func(string) string
}

// UpdateStructDescriptor modifies the field naming for struct serialization.
//
// It converts exported field names (starting with uppercase letter) to camelCase,
// unless they have an explicit JSON tag.
func (extension *namingStrategyExtension) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		if unicode.IsLower(rune(binding.Field.Name()[0])) || binding.Field.Name()[0] == '_' {
			continue
		}
		tag, hastag := binding.Field.Tag().Lookup("json")
		if hastag {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue // hidden field
			}
			if tagParts[0] != "" {
				continue // field explicitly named
			}
		}
		binding.ToNames = []string{extension.translate(binding.Field.Name())}
		binding.FromNames = []string{extension.translate(binding.Field.Name())}
	}
}

// lowerCaseFirstChar converts the first character of a string to lowercase.
//
// It is used for camelCase conversion (e.g., "FirstName" → "firstName").
func lowerCaseFirstChar(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return ""
}
