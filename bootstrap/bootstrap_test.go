package bootstrap_test

import (
	"os"
	"testing"

	"github.com/integration-system/isp-kit/bootstrap"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/stretchr/testify/require"
)

type RemoteConfig struct {
	SomeKey    string `validate:"required"`
	DynamicMap map[string]int
	Slice      []string
}

func TestNew(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	err := os.Setenv("APP_CONFIG_PATH", "test_data/config_test.yml")
	require.NoError(err)

	err = os.Setenv("DefaultRemoteConfigPath", "test_data/default_remote_config_test.json")
	require.NoError(err)

	_ = bootstrap.New("test", RemoteConfig{}, []cluster.EndpointDescriptor{{
		Path:             "test/endpoint",
		Inner:            true,
		UserAuthRequired: false,
		Extra:            cluster.RequireAdminPermission("perm"),
		Handler:          nil,
	}})
}
