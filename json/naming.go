package json

import (
	"strings"
	"unicode"

	"github.com/json-iterator/go"
)

type namingStrategyExtension struct {
	jsoniter.DummyExtension
	translate func(string) string
}

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

func lowerCaseFirstChar(s string) string {
	for i, v := range s {
		return string(unicode.ToLower(v)) + s[i+1:]
	}
	return ""
}
