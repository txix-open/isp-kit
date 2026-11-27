package rc_test

import (
	"strings"
	"testing"

	"github.com/txix-open/isp-kit/rc"
)

func TestTransformStringLeavesPaths(t *testing.T) {
	t.Parallel()

	out, err := rc.TransformStringLeaves(
		[]byte(`{"a":"x","nested":{"b":"y"},"list":["z","w"]}`),
		func(path, value string) string { return path + ":" + value },
	)
	if err != nil {
		t.Fatalf("TransformStringLeaves: %v", err)
	}

	want := `{"a":"$.a:x","list":["$.list[]:z","$.list[]:w"],"nested":{"b":"$.nested.b:y"}}`
	if string(out) != want {
		t.Fatalf("out = %s, want %s", out, want)
	}
}

func TestTransformNonStringLeaves(t *testing.T) {
	t.Parallel()

	var calls int
	var seenPath, seenValue string
	out, err := rc.TransformStringLeaves(
		[]byte(`{"n":5,"b":true,"z":null,"s":"x"}`),
		func(path, value string) string {
			calls++
			seenPath = path
			seenValue = value
			return path + ":" + value
		},
	)
	if err != nil {
		t.Fatalf("TransformStringLeaves: %v", err)
	}

	if calls != 1 || seenPath != "$.s" || seenValue != "x" {
		t.Fatalf("fn called %d times, last seen path=%q value=%q, want one call with $.s/x", calls, seenPath, seenValue)
	}
	want := `{"b":true,"n":5,"s":"$.s:x","z":null}`
	if string(out) != want {
		t.Fatalf("out = %s, want %s", out, want)
	}
}

func TestTransformKeysUntouched(t *testing.T) {
	t.Parallel()

	out, err := rc.TransformStringLeaves(
		[]byte(`{"_Var(k)":"v"}`),
		func(path, value string) string { return path + ":" + value },
	)
	if err != nil {
		t.Fatalf("TransformStringLeaves: %v", err)
	}

	want := `{"_Var(k)":"$._Var(k):v"}`
	if string(out) != want {
		t.Fatalf("out = %s, want %s", out, want)
	}
}

func TestTransformInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := rc.TransformStringLeaves(
		[]byte(`{"a":`),
		func(path, value string) string { return value },
	)
	if err == nil || !strings.Contains(err.Error(), "unmarshal config value") {
		t.Fatalf("err = %v, want an error containing %q", err, "unmarshal config value")
	}
}

func TestTransformEscapes(t *testing.T) {
	t.Parallel()

	var seen string
	out, err := rc.TransformStringLeaves(
		[]byte(`{"a":"line1\nline2"}`),
		func(path, value string) string {
			seen = value
			return strings.ReplaceAll(value, "\n", " | ")
		},
	)
	if err != nil {
		t.Fatalf("TransformStringLeaves: %v", err)
	}

	if seen != "line1\nline2" {
		t.Fatalf("fn saw %q, want the unescaped value %q", seen, "line1\nline2")
	}
	want := `{"a":"line1 | line2"}`
	if string(out) != want {
		t.Fatalf("out = %s, want %s", out, want)
	}
}
