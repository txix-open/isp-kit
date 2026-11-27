package schema_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/txix-open/isp-kit/rc/schema"
)

type uiExposedSetting struct {
	LoginEndpoint  string `schemaGen:"uiExposed" schema:"login endpoint"`
	LogoutEndpoint string `schema:"logout endpoint"`
}

type uiExposedClient struct {
	ClientId     string            `schemaGen:"uiExposed" schema:"client id"`
	ClientSecret string            `schema:"client secret"`
	Setting      *uiExposedSetting `schema:"settings"`
}

type uiExposedOauth struct {
	Host    string            `schema:"host"`
	Clients []uiExposedClient `schema:"clients"`
}

type uiExposedTestConfig struct {
	Secrets string          `schemaGen:"uiExposed" schema:"secrets"`
	Plain   string          `schema:"plain"`
	Oauth   *uiExposedOauth `schema:"oauth"`
	Extras  string          `jsonschema_extras:"x-foo=bar" schemaGen:"uiExposed" schema:"extras"`
}

func TestUIExposedGenerator(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(schema.NewGenerator().Generate(&uiExposedTestConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	var schemaDoc map[string]any
	err = json.Unmarshal(data, &schemaDoc)
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	leaf, ok := schemaDoc["properties"].(map[string]any)["secrets"].(map[string]any)
	if !ok {
		t.Fatalf("secrets leaf not found in schema: %s", data)
	}
	if leaf["x-ui-exposed"] != true {
		t.Fatalf("secrets leaf is not marked: %v", leaf)
	}

	// The keyword must not clobber other extras of the same field.
	extras, ok := schemaDoc["properties"].(map[string]any)["extras"].(map[string]any)
	if !ok {
		t.Fatalf("extras leaf not found in schema: %s", data)
	}
	if extras["x-ui-exposed"] != true || extras["x-foo"] != "bar" {
		t.Fatalf("extras leaf lost its other extras: %v", extras)
	}

	plain, ok := schemaDoc["properties"].(map[string]any)["plain"].(map[string]any)
	if !ok {
		t.Fatalf("plain leaf not found in schema: %s", data)
	}
	if _, marked := plain["x-ui-exposed"]; marked {
		t.Fatalf("plain leaf must not be marked: %v", plain)
	}
}

func TestKeywordReader(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(schema.NewGenerator().Generate(&uiExposedTestConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	paths, err := schema.UIExposedPaths(data)
	if err != nil {
		t.Fatalf("UIExposedPaths: %v", err)
	}

	want := []string{
		"$.extras",
		"$.oauth.clients[].clientId",
		"$.oauth.clients[].setting.loginEndpoint",
		"$.secrets",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}

func TestKeywordReaderUnmarked(t *testing.T) {
	t.Parallel()

	type plain struct {
		Value string `schema:"value"`
	}
	data, err := json.Marshal(schema.NewGenerator().Generate(&plain{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	paths, err := schema.UIExposedPaths(data)
	if err != nil {
		t.Fatalf("UIExposedPaths: %v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("paths = %v, want empty", paths)
	}
}

// A uiExposed tag on a string slice marks the array node; the elements are
// the exposed values, so the canonical path must match PlaceHolderReader's
// $.arr[] element paths.
func TestKeywordReaderStringArray(t *testing.T) {
	t.Parallel()

	type arr struct {
		Tags []string `schemaGen:"uiExposed" schema:"tags"`
	}
	data, err := json.Marshal(schema.NewGenerator().Generate(&arr{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	paths, err := schema.UIExposedPaths(data)
	if err != nil {
		t.Fatalf("UIExposedPaths: %v", err)
	}
	want := []string{"$.tags[]"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}

type sensitiveClient struct {
	Secret string `schemaGen:"sensitive" schema:"secret"`
	Host   string `schemaGen:"uiExposed" schema:"host"`
}

type sensitiveOauth struct {
	Clients []sensitiveClient `schema:"clients"`
}

type sensitiveTestConfig struct {
	Token  string          `schemaGen:"sensitive" schema:"token"`
	Login  string          `schemaGen:"uiExposed" schema:"login"`
	Tokens []string        `schemaGen:"sensitive" schema:"tokens"`
	Plain  string          `schema:"plain"`
	Oauth  *sensitiveOauth `schema:"oauth"`
}

func TestSensitiveGenerator(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(schema.NewGenerator().Generate(&sensitiveTestConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	props := schemaProperties(t, data)

	token, ok := props["token"].(map[string]any)
	if !ok {
		t.Fatalf("token leaf not found in schema: %s", data)
	}
	if token[schema.SensitiveKeyword] != true {
		t.Fatalf("token leaf is not marked: %v", token)
	}
	if _, exposed := token[schema.UIExposedKeyword]; exposed {
		t.Fatalf("token leaf must not carry the ui-exposed keyword: %v", token)
	}

	plain, ok := props["plain"].(map[string]any)
	if !ok {
		t.Fatalf("plain leaf not found in schema: %s", data)
	}
	if _, marked := plain[schema.SensitiveKeyword]; marked {
		t.Fatalf("plain leaf must not be marked: %v", plain)
	}
}

func TestSensitivePaths(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(schema.NewGenerator().Generate(&sensitiveTestConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	paths, err := schema.SensitivePaths(data)
	if err != nil {
		t.Fatalf("SensitivePaths: %v", err)
	}

	want := []string{
		"$.oauth.clients[].secret",
		"$.token",
		"$.tokens[]",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths = %v, want %v", paths, want)
	}
}

// The generic reader is keyed by the keyword: a schema carrying both marks
// must yield each wrapper only its own paths, with no cross-leak.
func TestKeywordReaderIsolation(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(schema.NewGenerator().Generate(&sensitiveTestConfig{}))
	if err != nil {
		t.Fatalf("marshal schema: %v", err)
	}

	uiPaths, err := schema.UIExposedPaths(data)
	if err != nil {
		t.Fatalf("UIExposedPaths: %v", err)
	}
	wantUI := []string{
		"$.login",
		"$.oauth.clients[].host",
	}
	if !reflect.DeepEqual(uiPaths, wantUI) {
		t.Fatalf("ui-exposed paths = %v, want %v", uiPaths, wantUI)
	}

	sensitivePaths, err := schema.SensitivePaths(data)
	if err != nil {
		t.Fatalf("SensitivePaths: %v", err)
	}
	wantSensitive := []string{
		"$.oauth.clients[].secret",
		"$.token",
		"$.tokens[]",
	}
	if !reflect.DeepEqual(sensitivePaths, wantSensitive) {
		t.Fatalf("sensitive paths = %v, want %v", sensitivePaths, wantSensitive)
	}
}

// schemaProperties unmarshals a generated schema and returns its top-level
// properties map.
func schemaProperties(t *testing.T, data []byte) map[string]any {
	t.Helper()

	var schemaDoc map[string]any
	err := json.Unmarshal(data, &schemaDoc)
	if err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	props, ok := schemaDoc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties: %s", data)
	}
	return props
}
