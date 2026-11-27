package rc

import (
	"encoding/json"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/rc/schema"
)

// TransformStringLeaves walks the JSON document and applies fn to every
// string leaf, passing its canonical path (see schema.CanonicalPath and
// schema.ArraySegments) and the unescaped value; the returned value replaces
// the leaf. JSON keys, non-string leaves and the document structure are
// untouched. The document is re-serialized, so object key order is not
// preserved.
func TransformStringLeaves(input []byte, fn func(path, value string) string) ([]byte, error) {
	var root any
	err := json.Unmarshal(input, &root)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config value")
	}

	root = transformLeaves(root, nil, fn)

	out, err := json.Marshal(root)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal config value")
	}
	return out, nil
}

func transformLeaves(node any, path []string, fn func(path, value string) string) any {
	switch v := node.(type) {
	case map[string]any:
		for key, child := range v {
			v[key] = transformLeaves(child, append(path, key), fn)
		}
		return v
	case []any:
		for i, item := range v {
			v[i] = transformLeaves(item, schema.ArraySegments(path), fn)
		}
		return v
	case string:
		return fn(schema.CanonicalPath(path), v)
	default:
		return node
	}
}
