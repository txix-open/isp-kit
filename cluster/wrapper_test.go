package cluster_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
)

const ExpectData = `
{
	"secret": {
		"1": "***",
		"2": "***"
	},
	"db": {
		"database": "dbname",
		"password": "***",
		"username": "dbname",
		"host": "101.00.11.11",
		"port": 5433
	},
	"rabbit": {
		"client": {
			"password": "***",
			"username": "username",
			"host": "102.00.22.22",
			"port": 5673
		}
	},
	"Test": {
		"time": "80h",
		"Secret": ""
	},
	"token":"***" 
}
`

const ConfigData = `
{
	"secret": {
		"1": "***",
		"2": "***"
	},
	"db": {
		"database": "dbname",
		"password": "***",
		"username": "dbname",
		"host": "101.00.11.11",
		"port": 5433
	},
	"rabbit": {
		"client": {
			"password": "***",
			"username": "username",
			"host": "102.00.22.22",
			"port": 5673
		}
	},
	"Test": {
		"time": "80h",
		"Secret": ""
	},
	"token":"***" 
}
`

func TestHideSecrets(t *testing.T) {
	require := require.New(t)
	configData := []byte(ConfigData)
	expectData := []byte(ExpectData)

	actualData, err := cluster.HideSecrets(configData)
	require.NoError(err)

	expectConfig := make(map[string]any)
	actualConfig := make(map[string]any)

	err = json.Unmarshal(expectData, &expectConfig)
	require.NoError(err)
	err = json.Unmarshal(actualData, &actualConfig)
	require.NoError(err)

	require.Equal(expectConfig, actualConfig)
}
