// nolint:goconst,mnd
package schema

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/txix-open/jsonschema"
)

const (
	tagDefault      = "default"
	tagSchema       = "schema"
	tagCustomSchema = "schemaGen"
)

// GetNameAndRequiredFlag extracts the JSON field name and determines if the field is required.
// It checks the struct field's json tag for the name and validate tag for the "required" constraint.
// Returns the field name and a boolean indicating if the field is required.
// Unexported fields (those with non-empty PkgPath) are ignored.
func GetNameAndRequiredFlag(field reflect.StructField) (string, bool) {
	if field.PkgPath != "" { // unexported field, ignore it
		return "", false
	}

	name, accept := GetFieldName(field)
	if !accept {
		return "", false
	}

	validators := getValidatorsMap(field)
	if validators != nil {
		_, present := validators["required"]
		return name, present
	}

	return name, false
}

// SetProperties populates a JSON schema with properties from struct field tags.
// It processes the following tags:
//   - "schema": sets Title and Description (format: "title,description")
//   - "default": sets the Default value
//   - "validate": sets validation constraints (min, max, oneof, etc.)
//   - "schemaGen": applies a custom generator by name
func SetProperties(field reflect.StructField, s *jsonschema.Schema) {
	schema, ok := field.Tag.Lookup(tagSchema)
	if ok {
		parts := strings.SplitN(schema, ",", 2)
		if len(parts) > 0 {
			if len(parts) == 2 {
				s.Description = parts[1]
			}
			if parts[0] != "" {
				s.Title = parts[0]
			}
		}
	}

	defaultValue, ok := field.Tag.Lookup(tagDefault)
	if ok {
		s.Default = defaultValue
	}

	setValidators(field, s)

	customValue, ok := field.Tag.Lookup(tagCustomSchema)
	if ok {
		if f := CustomGenerators.getGeneratorByName(customValue); f != nil {
			f(field, s)
		}
	}
}

// setValidators applies validation constraints from struct field tags to the schema.
// Supported constraints: max/lte, min/gte, oneof.
func setValidators(field reflect.StructField, s *jsonschema.Schema) {
	validators := getValidatorsMap(field)
	if validators == nil {
		return
	}
	for k, v := range validators {
		switch k {
		case "max", "lte":
			setMax(s, v)
		case "min", "gte":
			setMin(s, v)
		case "oneof":
			setOneOf(s, v)
		}
	}
}

// setMax sets the maximum constraint on a schema based on type.
// For strings, it sets MaxLength; for integers, Maximum; for arrays, MaxItems.
// Returns early if the value cannot be parsed as an unsigned integer.
func setMax(s *jsonschema.Schema, v string) {
	value, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return
	}
	switch s.Type {
	case "string":
		s.MaxLength = &value
	case "integer":
		s.Maximum = json.Number(v)
	case "array":
		s.MaxItems = &value
	}
}

// setMin sets the minimum constraint on a schema based on type.
// For strings, it sets MinLength; for integers, Minimum; for arrays, MinItems.
// Returns early if the value cannot be parsed as an unsigned integer.
func setMin(s *jsonschema.Schema, v string) {
	value, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return
	}
	switch s.Type {
	case "string":
		s.MinLength = &value
	case "integer":
		s.Minimum = json.Number(v)
	case "array":
		s.MinItems = &value
	}
}

// setOneOf sets an Enum list on the schema for string or integer types.
// Returns early if the schema type is not string or integer.
func setOneOf(s *jsonschema.Schema, v string) {
	if s.Type != "string" && s.Type != "integer" {
		return
	}
	for _, value := range parseOneOfParam(v) {
		s.Enum = append(s.Enum, value)
	}
}

// getValidatorsMap extracts validation constraints from the "validate" struct tag.
// Returns a map of constraint names to their values, or nil if no validate tag exists.
func getValidatorsMap(field reflect.StructField) tagOptionsMap {
	value, ok := field.Tag.Lookup("validate")
	if !ok {
		return nil
	}
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return nil
	}
	return parseTagIntoMap(value)
}

// tagOptionsMap represents a map of tag option names to their values.
type tagOptionsMap map[string]string

// parseTagIntoMap converts a comma-separated tag string into a map.
// Supports both key-only (e.g., "required") and key=value (e.g., "oneof='a','b'") formats.
func parseTagIntoMap(tag string) tagOptionsMap {
	optionsMap := make(tagOptionsMap)
	for option := range strings.SplitSeq(tag, ",") {
		option = strings.TrimSpace(option)
		keyValue := strings.Split(option, "=")
		switch len(keyValue) {
		case 1:
			optionsMap[keyValue[0]] = ""
		case 2:
			optionsMap[keyValue[0]] = keyValue[1]
		}
	}
	return optionsMap
}

// GetFieldName extracts the JSON field name from a struct field.
// It checks the "json" tag for an explicit name; if absent, it uses the field name
// converted to camelCase (first letter lowercase).
// Returns the field name and true if the field should be included, or empty string and false if not.
func GetFieldName(fieldType reflect.StructField) (string, bool) {
	original := fieldType.Name
	transform := true
	value, ok := fieldType.Tag.Lookup("json")
	if ok {
		opts := strings.Split(value, ",")
		if len(opts) > 0 {
			if opts[0] == "-" {
				return "", false
			}
			name := opts[0]
			if len(name) > 0 {
				original = name
				transform = false
			} else {
				transform = true
			}
		}
	}
	if transform {
		arr := []rune(original)
		arr[0] = unicode.ToLower(arr[0])
		original = string(arr)
	}
	return original, true
}

var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)

// parseOneOfParam parses a oneOf validation parameter string.
// It handles both quoted values (e.g., "'a','b','c'") and unquoted values.
// Returns a slice of parsed values with quotes removed.
func parseOneOfParam(param string) []string {
	values := splitParamsRegex.FindAllString(param, -1)
	for i := range values {
		values[i] = strings.ReplaceAll(values[i], "'", "")
	}
	return values
}
