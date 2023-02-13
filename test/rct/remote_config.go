package rct

import (
	"os"
	"testing"

	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/rc/schema"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func Test(t *testing.T, defaultRemoteConfigPath string, remoteConfig any) {
	require := require.New(t)

	defaultRemoteConfig, err := os.ReadFile(defaultRemoteConfigPath)
	require.NoError(err)

	jsonSchema := schema.GenerateConfigSchema(remoteConfig)
	jsonSchemaData, err := json.Marshal(jsonSchema)
	require.NoError(err)

	err = json.Unmarshal(defaultRemoteConfig, &remoteConfig)
	require.NoError(err)

	schemaLoader := gojsonschema.NewBytesLoader(jsonSchemaData)
	configLoader := gojsonschema.NewGoLoader(remoteConfig)
	result, err := gojsonschema.Validate(schemaLoader, configLoader)
	require.NoError(err)

	for _, resultError := range result.Errors() {
		require.Empty(resultError.String())
	}
}
