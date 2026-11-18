package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
)

const oldModuleReadyPayload = `{
	"moduleName": "isp-test-service",
	"version": "1.0.0",
	"libVersion": "1.75.0",
	"transport": "grpc",
	"endpoints": [{"path": "test/endpoint", "inner": true, "userAuthRequired": false, "httpMethod": ""}],
	"requiredModules": [{"name": "isp-auth-service", "required": true}],
	"address": {"ip": "10.0.0.1", "port": "9003"},
	"metricsAutodiscovery": {"address": "10.0.0.1:9553", "labels": {"app": "isp-test-service"}}
}`

type oldAddressConfiguration struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type oldBackendDeclaration struct {
	ModuleName           string
	Version              string
	LibVersion           string
	Transport            string
	Endpoints            []cluster.EndpointDescriptor
	RequiredModules      []cluster.ModuleDependency
	Address              oldAddressConfiguration
	MetricsAutodiscovery *cluster.MetricsAutodiscovery
}

func TestModuleReadyPayload_OldAnnouncementDecodesWithNewContract(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	var declaration cluster.BackendDeclaration
	err := json.Unmarshal([]byte(oldModuleReadyPayload), &declaration)
	require.NoError(err)

	require.Equal("isp-test-service", declaration.ModuleName)
	require.Equal("1.0.0", declaration.Version)
	require.Equal("1.75.0", declaration.LibVersion)
	require.Equal(cluster.GrpcTransport, declaration.Transport)
	require.Equal([]cluster.ModuleDependency{{Name: "isp-auth-service", Required: true}}, declaration.RequiredModules)
	require.Equal("10.0.0.1", declaration.Address.IP)
	require.Equal("9003", declaration.Address.Port)
	require.Equal("10.0.0.1:9553", declaration.MetricsAutodiscovery.Address)
	require.Equal(cluster.ModuleLeaderModeNone, declaration.ModuleLeaderMode)
	require.False(declaration.Address.IsModuleLeader)
}

func TestModuleReadyPayload_NewAnnouncementDecodesWithOldContract(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	declaration := cluster.BackendDeclaration{
		ModuleName: "isp-test-service",
		Version:    "1.0.0",
		LibVersion: "1.76.0",
		Transport:  cluster.GrpcTransport,
		Endpoints: []cluster.EndpointDescriptor{{
			Path:  "test/endpoint",
			Inner: true,
		}},
		RequiredModules: []cluster.ModuleDependency{{Name: "isp-auth-service", Required: true}},
		Address: cluster.AddressConfiguration{
			IP:             "10.0.0.1",
			Port:           "9003",
			IsModuleLeader: true,
		},
		ModuleLeaderMode: cluster.ModuleLeaderModeRequestElection,
	}

	payload, err := json.Marshal(declaration)
	require.NoError(err)

	wire := string(payload)
	require.Contains(wire, `"moduleLeaderMode":"REQUEST_ELECTION"`)
	require.Contains(wire, `"isModuleLeader":true`)

	var old oldBackendDeclaration
	err = json.Unmarshal(payload, &old)
	require.NoError(err)

	require.Equal(declaration.ModuleName, old.ModuleName)
	require.Equal(declaration.Version, old.Version)
	require.Equal(declaration.LibVersion, old.LibVersion)
	require.Equal(declaration.Transport, old.Transport)
	require.Equal(declaration.Endpoints, old.Endpoints)
	require.Equal(declaration.RequiredModules, old.RequiredModules)
	require.Equal(oldAddressConfiguration{IP: "10.0.0.1", Port: "9003"}, old.Address)
}

func TestAddressPayload_StandbyLeaderFlagIsExplicit(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	connectedPayload, err := json.Marshal([]cluster.AddressConfiguration{{IP: "10.0.0.1", Port: "9003"}})
	require.NoError(err)
	require.Contains(string(connectedPayload), `"isModuleLeader":false`)

	routesPayload, err := json.Marshal(cluster.RoutingConfig{{
		ModuleName: "isp-test-service",
		Transport:  cluster.HttpTransport,
		Address:    cluster.AddressConfiguration{IP: "10.0.0.1", Port: "9003"},
	}})
	require.NoError(err)
	require.Contains(string(routesPayload), `"isModuleLeader":false`)
	require.Contains(string(routesPayload), `"moduleLeaderMode":""`)
}
