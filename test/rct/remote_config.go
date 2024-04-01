package rct

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/validator"
	"github.com/xeipuuv/gojsonschema"
)

func Test[T any](t *testing.T, defaultRemoteConfigPath string, remoteConfig T) {
	require := require.New(t)

	validatorTag := "valid"
	require.Falsef(FindTag(remoteConfig, validatorTag), "'%s' tag is not longer supported, you must use https://github.com/go-playground/validator", validatorTag)

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

func FindTag[T any](v T, tag string) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}
	queue := []reflect.Type{t}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur.Kind() == reflect.Ptr {
			cur = cur.Elem()
		}
		switch cur.Kind() {
		case reflect.Struct:
			for i := 0; i < cur.NumField(); i++ {
				field := cur.Field(i)
				if field.Tag.Get(tag) != "" {
					return true
				}
				queue = append(queue, field.Type)
			}
		case reflect.Map, reflect.Array, reflect.Slice:
			queue = append(queue, cur.Elem())
		}
	}
	return false
}
