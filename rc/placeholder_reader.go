package rc

import (
	"encoding/json"
	"regexp"
	"strings"
)

var variableRefPattern = regexp.MustCompile(`_Var\((.*?)\)`)

// IsCompositeJSONString reports whether value is a JSON-encoded composite
// value: a non-empty string that, trimmed of surrounding whitespace, starts
// with '{' or '[' and unmarshals to a JSON object or array. JSON scalars
// ("42", "\"text\"") and malformed JSON are not composite.
func IsCompositeJSONString(value string) bool {
	trimmed := strings.TrimLeft(value, " \t\r\n")
	if len(trimmed) == 0 {
		return false
	}
	if trimmed[0] != '{' && trimmed[0] != '[' {
		return false
	}
	var node any
	err := json.Unmarshal([]byte(trimmed), &node)
	if err != nil {
		return false
	}
	switch node.(type) {
	case map[string]any, []any:
		return true
	default:
		return false
	}
}

// collectVariableRefs collects the canonical paths of the config value's
// string leaves that contain _Var(name) references, mapped to the referenced
// names (de-duplicated, in order of first occurrence). When skipComposite is
// true, composite JSON string leaves are ignored.
func collectVariableRefs(data []byte, skipComposite bool) (map[string][]string, error) {
	refs := make(map[string][]string)
	_, err := TransformStringLeaves(data, func(path, value string) string {
		if skipComposite && IsCompositeJSONString(value) {
			return value
		}
		matches := variableRefPattern.FindAllStringSubmatch(value, -1)
		if len(matches) == 0 {
			return value
		}
		names := make([]string, 0, len(matches))
		seen := make(map[string]struct{}, len(matches))
		for _, m := range matches {
			name := m[1]
			if _, dup := seen[name]; dup {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
		refs[path] = append(refs[path], names...)
		return value
	})
	return refs, err
}

func PlaceHolderReader(data []byte) (map[string][]string, error) {
	return collectVariableRefs(data, true)
}

func VariableRefReport(data []byte) (map[string][]string, error) {
	return collectVariableRefs(data, false)
}

// SingleVariableRef reports whether value consists of exactly one _Var(name)
// reference with no surrounding text, and returns the referenced name.
// Mixed values ("prefix _Var(a)"), multiple references and plain strings are
// not single references.
func SingleVariableRef(value string) (string, bool) {
	match := variableRefPattern.FindStringSubmatch(value)
	if match == nil || match[0] != value {
		return "", false
	}
	return match[1], true
}
