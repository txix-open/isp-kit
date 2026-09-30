// Package rct provides test helpers for remote configuration validation.
// It validates that configuration structures match their JSON schemas and
// can be properly unmarshaled from configuration files.
package rct

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/validator"
	"github.com/xeipuuv/gojsonschema"
)

// Test validates that the provided configuration structure matches its
// generated JSON schema and can be properly unmarshaled from the default
// configuration file. It also verifies that the configuration passes
// validator tag validation.
//
// The function checks:
//   - The default configuration file is valid against the generated schema
//   - The configuration can be unmarshaled from the file
//   - The configuration passes validation
//
// nolint:lll
func Test[T any](t *testing.T, defaultRemoteConfigPath string, remoteConfig T) {
	require := require.New(t)

	defaultRemoteConfig, err := os.ReadFile(defaultRemoteConfigPath)
	require.NoError(err)

	jsonSchema := rc.GenerateConfigSchema(remoteConfig)
	jsonSchemaData, err := json.Marshal(jsonSchema)
	require.NoError(err)

	remoteConfigAsMap := make(map[string]any)
	err = json.Unmarshal(defaultRemoteConfig, &remoteConfigAsMap)
	require.NoError(err)

	schemaLoader := gojsonschema.NewBytesLoader(jsonSchemaData)
	configLoader := gojsonschema.NewGoLoader(remoteConfigAsMap)
	result, err := gojsonschema.Validate(schemaLoader, configLoader)
	require.NoError(err)

	for _, resultError := range result.Errors() {
		require.Empty(resultError.String())
	}

	err = json.Unmarshal(defaultRemoteConfig, &remoteConfig)
	require.NoError(err)
	err = validator.Default.ValidateToError(remoteConfig)
	require.NoError(err)
}
