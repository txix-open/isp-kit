package schema

import (
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/asaskevich/govalidator"
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

	if validators := getValidatorsMap(field); validators != nil {
		_, present := validators["required"]
		return name, present
	}

	return name, false
}

func SetProperties(field reflect.StructField, t *jsonschema.Type) {
	schema, ok := field.Tag.Lookup(tagSchema)
	if ok {
		parts := strings.SplitN(schema, ",", 2)
		if len(parts) > 0 {
			if len(parts) == 2 {
				t.Description = parts[1]
			}
			if parts[0] != "" {
				t.Title = parts[0]
			}
		}
	}

	if defaultValue, ok := field.Tag.Lookup(tagDefault); ok {
		t.Default = defaultValue
	}

	setValidators(field, t)

	if customValue, ok := field.Tag.Lookup(tagCustomSchema); ok {
		if f := CustomGenerators.getGeneratorByName(customValue); f != nil {
			f(field, t)
		}
	}
}

func setValidators(field reflect.StructField, t *jsonschema.Type) {
	validators := getValidatorsMap(field)
	if validators == nil {
		return
	}

	for f, args := range validators {
		switch f {
		case "uri":
			fallthrough
		case "email":
			fallthrough
		case "ipv4":
			fallthrough
		case "ipv6":
			t.Format = f
		case "matches":
			if len(args) > 0 {
				t.Pattern = args[0]
			}
		case "host":
			t.Format = "hostname"
		case "length":
			fallthrough
		case "runelength":
			if len(args) > 0 {
				if val, err := strconv.Atoi(args[0]); err == nil {
					t.MinLength = val
				}
				if len(args) > 1 {
					if val, err := strconv.Atoi(args[1]); err == nil {
						t.MaxLength = val
					}
				}
			}
		case "in":
			if len(args) > 0 {
				vals := strings.Split(args[0], "|")
				for _, val := range vals {
					t.Enum = append(t.Enum, val)
				}
			}
		case "range":
			if len(args) > 0 {
				if val, err := strconv.Atoi(args[0]); err == nil {
					t.Minimum = val
				}
				if len(args) > 1 {
					if val, err := strconv.Atoi(args[1]); err == nil {
						t.Maximum = val
					}
				}
			}
		}
	}
}

func getValidatorsMap(field reflect.StructField) map[string][]string {
	value, ok := field.Tag.Lookup("valid")
	if !ok {
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return nil
	}

	tm := parseTagIntoMap(value)
	result := make(map[string][]string, len(tm))
	for val := range tm {
		f, args := getValidatorFunction(val)
		result[f] = args
	}
	return result
}

func getValidatorFunction(val string) (string, []string) {
	for key, value := range govalidator.ParamTagRegexMap {
		ps := value.FindStringSubmatch(val)
		l := len(ps)
		if l < 2 {
			continue
		}
		args := make([]string, l-1)
		for i := 1; i < l; i++ {
			args[i-1] = ps[i]
		}
		return key, args
	}
	return val, []string{}
}

type tagOptionsMap map[string]string

func parseTagIntoMap(tag string) tagOptionsMap {
	optionsMap := make(tagOptionsMap)
	options := strings.Split(tag, ",")

	for _, option := range options {
		option = strings.TrimSpace(option)

		validationOptions := strings.Split(option, "~")
		if !isValidTag(validationOptions[0]) {
			continue
		}
		if len(validationOptions) == 2 {
			optionsMap[validationOptions[0]] = validationOptions[1]
		} else {
			optionsMap[validationOptions[0]] = ""
		}
	}
	return optionsMap
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

func GetFieldName(fieldType reflect.StructField) (string, bool) {
	original := fieldType.Name
	transform := true
	if value, ok := fieldType.Tag.Lookup("json"); ok {
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
