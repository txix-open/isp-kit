package schema

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/txix-open/isp-kit/errors"
)

func UIExposedPaths(data []byte) ([]string, error) { return keywordReader(data, UIExposedKeyword) }
func SensitivePaths(data []byte) ([]string, error) { return keywordReader(data, SensitiveKeyword) }

func keywordReader(data []byte, keyword string) ([]string, error) {
	var root any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, errors.WithMessage(err, "unmarshal schema")
	}

	paths := make([]string, 0)
	collectTagged(root, nil, &paths, keyword)
	sort.Strings(paths)

	return paths, nil
}

func collectTagged(node any, path []string, paths *[]string, keyword string) {
	obj, ok := node.(map[string]any)
	if !ok {
		return
	}

	if _, marked := obj[keyword]; marked {
		p := path
		if _, isArray := obj["items"]; isArray {
			p = ArraySegments(path)
		}
		*paths = append(*paths, CanonicalPath(p))
		return
	}

	if props, ok := obj["properties"].(map[string]any); ok {
		for name, child := range props {
			collectTagged(child, append(path, name), paths, keyword)
		}
	}

	switch items := obj["items"].(type) {
	case map[string]any:
		collectTagged(items, ArraySegments(path), paths, keyword)
	case []any:
		for _, item := range items {
			collectTagged(item, ArraySegments(path), paths, keyword)
		}
	}
}

func CanonicalPath(segments []string) string {
	return "$." + strings.Join(segments, ".")
}

func ArraySegments(path []string) []string {
	if len(path) == 0 {
		return append(path, "[]")
	}
	last := path[len(path)-1] + "[]"
	return append(path[:len(path)-1:len(path)-1], last)
}
