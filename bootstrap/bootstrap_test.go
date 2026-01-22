package bootstrap_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
)

type RemoteConfig struct {
	SomeKey    string `validate:"required"`
	DynamicMap map[string]int
	Slice      []string
}

type TestAppConfig struct {
	Field1 string
	Field2 int
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
	}}, cluster.GrpcTransport)
}

type mockConfigReceiver struct {
	t                    *testing.T
	expectedRemoteConfig []byte
	callCount            int
}

func (m *mockConfigReceiver) ReceiveConfig(_ context.Context, remoteConfig []byte) error {
	assert.Equal(m.t, m.expectedRemoteConfig, remoteConfig)
	m.callCount++
	return nil
}

func TestNew_Offline(t *testing.T) {
	t.Setenv("CLUSTER_MODE", "offline")
	t.Setenv("APP_CONFIG_PATH", "test_data/config_offline_test.yml")

	boot := bootstrap.New("test", RemoteConfig{}, []cluster.EndpointDescriptor{{
		Path:             "test/endpoint",
		Inner:            true,
		UserAuthRequired: false,
		Extra:            cluster.RequireAdminPermission("perm"),
		Handler:          nil,
	}}, cluster.GrpcTransport)

	expectedRemoteConfig, err := os.ReadFile("test_data/config_test.json")
	require.NoError(t, err)

	receiver := &mockConfigReceiver{t: t, expectedRemoteConfig: expectedRemoteConfig}
	handler := cluster.NewEventHandler().RemoteConfigReceiver(receiver)

	err = boot.ClusterCli.Run(t.Context(), handler)
	require.NoError(t, err)
	assert.Equal(t, 1, receiver.callCount)
}

func TestNewStandalone(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", "test_data/config_standalone_test.yml")

	boot := bootstrap.NewStandalone("test")

	var cfg TestAppConfig
	err := boot.ReadConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "field1_test_value", cfg.Field1)
	assert.Equal(t, 10, cfg.Field2)
}

func TestNewStandalone_InvalidConfig(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", "test_data/config_standalone_test.yml")

	boot := bootstrap.NewStandalone("test")

	type InvalidConfig struct {
		Field1 string `validate:"required"`
		Name   string `validate:"required"`
	}
	var cfg InvalidConfig
	err := boot.ReadConfig(&cfg)
	require.Error(t, err)
}
