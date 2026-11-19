package cluster_test

import (
	"context"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/etp/v4"
	"github.com/txix-open/etp/v4/msg"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

const moduleReadyWaitTimeout = 10 * time.Second

func TestModuleReadyAnnouncementCarriesModuleLeaderMode(t *testing.T) {
	t.Parallel()

	require := require.New(t)

	readyPayload := make(chan []byte, 1)
	ack := etp.HandlerFunc(func(_ context.Context, _ *etp.Conn, event msg.Event) []byte {
		if event.Name == cluster.ModuleReady {
			recordPayload(readyPayload, event.Data)
		}
		return []byte("ok")
	})
	configService := etp.NewServer()
	for _, event := range []string{cluster.ModuleSendConfigSchema, cluster.ModuleSendRequirements, cluster.ModuleReady} {
		configService.On(event, ack)
	}
	server := httptest.NewServer(configService)
	defer server.Close()

	host, port, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(err)
	configServiceAddress := net.JoinHostPort(host, port)

	moduleInfo := cluster.ModuleInfo{
		ModuleName:       "isp-test-service",
		ModuleVersion:    "1.0.0",
		LibVersion:       "1.76.0",
		Transport:        cluster.GrpcTransport,
		GrpcOuterAddress: cluster.AddressConfiguration{IP: host, Port: port},
		ModuleLeaderMode: cluster.ModuleLeaderModeRequestElection,
	}
	configData := cluster.ConfigData{
		Version: "1.0.0",
		Schema:  json.RawMessage(`{}`),
		Config:  json.RawMessage(`{}`),
	}

	logger, err := log.New(log.WithDisableDefaultOutput())
	require.NoError(err)

	client := cluster.NewClient(moduleInfo, configData, []string{configServiceAddress}, 0, logger)

	sessionCtx, cancelSession := context.WithCancel(t.Context())
	defer cancelSession()

	sessionErr := make(chan error, 1)
	go func() {
		sessionErr <- client.Run(sessionCtx, cluster.NewEventHandler())
	}()

	select {
	case payload := <-readyPayload:
		require.Contains(string(payload), `"moduleName":"isp-test-service"`)
		require.Contains(string(payload), `"moduleLeaderMode":"REQUEST_ELECTION"`)
		require.Contains(string(payload), `"isModuleLeader":false`)
	case err := <-sessionErr:
		require.NoError(err)
		t.Fatal("config service session stopped before MODULE:READY")
	case <-time.After(moduleReadyWaitTimeout):
		t.Fatal("timeout waiting for MODULE:READY")
	}

	require.NoError(client.Close())
}

func recordPayload(ch chan []byte, data []byte) {
	select {
	case ch <- data:
	default:
	}
}
