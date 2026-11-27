package rc_test

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/txix-open/isp-kit/rc"
)

func TestPlaceHolderReader(t *testing.T) {
	t.Parallel()

	data := []byte(`{
		"secrets": "_Var(secret_token)",
		"plain": "no references here",
		"mixed": "prefix/_Var(a) middle _Var(b) _Var(a)",
		"list": ["_Var(c)", "nope"],
		"nested": {"arr": [{"e": "_Var(d)"}, {"e": "x"}]}
	}`)

	refs, err := rc.PlaceHolderReader(data)
	if err != nil {
		t.Fatalf("PlaceHolderReader: %v", err)
	}

	if got := refs["$.secrets"]; !reflect.DeepEqual(got, []string{"secret_token"}) {
		t.Fatalf("$.secrets = %v, want [secret_token]", got)
	}

	// Contains (not prefix), de-duplicated, in order of first occurrence.
	if got := refs["$.mixed"]; !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("$.mixed = %v, want [a b]", got)
	}

	// Array elements collapse into one "[]" segment.
	if got := refs["$.list[]"]; !reflect.DeepEqual(got, []string{"c"}) {
		t.Fatalf("$.list[] = %v, want [c]", got)
	}

	// Nested arrays use the same currency as the schema reader.
	if got := refs["$.nested.arr[].e"]; !reflect.DeepEqual(got, []string{"d"}) {
		t.Fatalf("$.nested.arr[].e = %v, want [d]", got)
	}

	if _, ok := refs["$.plain"]; ok {
		t.Fatalf("$.plain must not be reported, got %v", refs)
	}

	// No spurious paths.
	if len(refs) != 4 {
		keys := make([]string, 0, len(refs))
		for k := range refs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		t.Fatalf("expected 4 paths, got %d: %v", len(refs), keys)
	}
}

func TestPlaceHolderReaderEmpty(t *testing.T) {
	t.Parallel()

	refs, err := rc.PlaceHolderReader([]byte(`{"a": "1", "b": {"c": []}}`))
	if err != nil {
		t.Fatalf("PlaceHolderReader: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("refs = %v, want empty", refs)
	}
}

func TestPlaceHolderReaderSkipsCompositeStrings(t *testing.T) {
	t.Parallel()

	// composite values are JSON documents stored as plain strings
	compositeObj, err := json.Marshal(map[string]any{"a": "_Var(secret_token)"})
	if err != nil {
		t.Fatalf("marshal composite object: %v", err)
	}
	compositeArr, err := json.Marshal([]any{"_Var(secret_token)"})
	if err != nil {
		t.Fatalf("marshal composite array: %v", err)
	}
	data, err := json.Marshal(map[string]any{
		"composite":     string(compositeObj),
		"compositeList": string(compositeArr),
		"plain":         "_Var(secret_token)",
		"list":          []any{"_Var(secret_token)"},
		"jsonScalar":    "42",
	})
	if err != nil {
		t.Fatalf("marshal test data: %v", err)
	}

	refs, err := rc.PlaceHolderReader(data)
	if err != nil {
		t.Fatalf("PlaceHolderReader: %v", err)
	}

	if got := refs["$.plain"]; !reflect.DeepEqual(got, []string{"secret_token"}) {
		t.Fatalf("$.plain = %v, want [secret_token]", got)
	}
	if got := refs["$.list[]"]; !reflect.DeepEqual(got, []string{"secret_token"}) {
		t.Fatalf("$.list[] = %v, want [secret_token]", got)
	}

	for _, path := range []string{"$.composite", "$.compositeList", "$.jsonScalar"} {
		if _, ok := refs[path]; ok {
			t.Fatalf("%s must not be reported, got %v", path, refs)
		}
	}
}

func TestVariableRefReportIncludesCompositeStrings(t *testing.T) {
	t.Parallel()

	compositeObj, err := json.Marshal(map[string]any{"a": "_Var(secret_token)"})
	if err != nil {
		t.Fatalf("marshal composite object: %v", err)
	}
	compositeArr, err := json.Marshal([]any{"_Var(secret_token)"})
	if err != nil {
		t.Fatalf("marshal composite array: %v", err)
	}
	data, err := json.Marshal(map[string]any{
		"composite":     string(compositeObj),
		"compositeList": string(compositeArr),
		"plain":         "_Var(secret_token)",
	})
	if err != nil {
		t.Fatalf("marshal test data: %v", err)
	}

	refs, err := rc.VariableRefReport(data)
	if err != nil {
		t.Fatalf("VariableRefReport: %v", err)
	}

	for _, path := range []string{"$.composite", "$.compositeList", "$.plain"} {
		if got := refs[path]; !reflect.DeepEqual(got, []string{"secret_token"}) {
			t.Fatalf("%s = %v, want [secret_token]", path, got)
		}
	}
	if len(refs) != 3 {
		keys := make([]string, 0, len(refs))
		for k := range refs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		t.Fatalf("expected 3 paths, got %d: %v", len(refs), keys)
	}
}

func TestSingleVariableRef(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value string
		name  string
		isRef bool
	}{
		"single ref":        {"_Var(a)", "a", true},
		"ref with space":    {"_Var(a) ", "", false},
		"ref with prefix":   {"x_Var(a)", "", false},
		"two refs":          {"_Var(a) _Var(b)", "", false},
		"two refs adjacent": {"_Var(a)_Var(b)", "", false},
		"no ref":            {"plain value", "", false},
		"empty":             {"", "", false},
		"nested parens":     {"_Var(a(b))", "", false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotName, gotRef := rc.SingleVariableRef(test.value)
			if gotRef != test.isRef || gotName != test.name {
				t.Fatalf("SingleVariableRef(%q) = (%q, %v), want (%q, %v)",
					test.value, gotName, gotRef, test.name, test.isRef)
			}
		})
	}
}

func TestIsCompositeJSONString(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value string
		want  bool
	}{
		"object":          {"{\"a\": 1}", true},
		"array":           {"[1, 2]", true},
		"object indented": {"  \n\t{\"a\": 1}", true},
		"empty object":    {"{}", true},
		"empty array":     {"[]", true},
		"empty string":    {"", false},
		"whitespace":      {"   ", false},
		"plain string":    {"_Var(name)", false},
		"json scalar":     {"42", false},
		"quoted scalar":   {"\"text\"", false},
		"malformed":       {"{\"a\": ", false},
		"brace tail":      {"{not json}", false},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := rc.IsCompositeJSONString(test.value); got != test.want {
				t.Fatalf("IsCompositeJSONString(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
