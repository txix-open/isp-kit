package bootstrap_test

import (
	"testing"

	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
)

type RemoteConfig struct {
	SomeKey    string `validate:"required"`
	DynamicMap map[string]int
	Slice      []string
}

func TestNew(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", "test_data/config_test.yml")
	t.Setenv("DefaultRemoteConfigPath", "test_data/default_remote_config_test.json")

	_ = bootstrap.New("test", RemoteConfig{}, []cluster.EndpointDescriptor{{
		Path:             "test/endpoint",
		Inner:            true,
		UserAuthRequired: false,
		Extra:            cluster.RequireAdminPermission("perm"),
		Handler:          nil,
	}})
}
