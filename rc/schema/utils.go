package schema

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/integration-system/jsonschema"
)

const (
	tagDefault      = "default"
	tagSchema       = "schema"
	tagCustomSchema = "schemaGen"
)

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

func setOneOf(s *jsonschema.Schema, v string) {
	if s.Type != "string" && s.Type != "integer" {
		return
	}
	for _, value := range parseOneOfParam(v) {
		s.Enum = append(s.Enum, value)
	}
}

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

type tagOptionsMap map[string]string

func parseTagIntoMap(tag string) tagOptionsMap {
	optionsMap := make(tagOptionsMap)
	options := strings.Split(tag, ",")
	for _, option := range options {
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

func parseOneOfParam(param string) []string {
	values := splitParamsRegex.FindAllString(param, -1)
	for i := 0; i < len(values); i++ {
		values[i] = strings.ReplaceAll(values[i], "'", "")
	}
	return values
}
