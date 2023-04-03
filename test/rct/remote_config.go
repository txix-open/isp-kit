package rct

import (
	"os"
	"testing"

	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/integration-system/isp-kit/validator"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func Test[T any](t *testing.T, defaultRemoteConfigPath string, remoteConfig T) {
	require := require.New(t)

	defaultRemoteConfig, err := os.ReadFile(defaultRemoteConfigPath)
	require.NoError(err)

	jsonSchema := schema.GenerateConfigSchema(remoteConfig)
	jsonSchemaData, err := json.Marshal(jsonSchema)
	require.NoError(err)

	remoteConfigAsMap := make(map[string]any, 0)
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
