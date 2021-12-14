package client_test

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/client"
	"github.com/integration-system/isp-kit/grpc/endpoint"
	"github.com/integration-system/isp-kit/log"
	"github.com/stretchr/testify/require"
)

func TestBalancing(t *testing.T) {
	require := require.New(t)

	servers := 3
	calls := 1500
	callsPerHost := int32(500)
	delta := int32(servers)

	hosts := make([]string, 0)
	callCounter := make([]int32, servers)
	for i := 0; i < servers; i++ {
		ptr := &callCounter[i]
		host := prepareServer(t, require, "test", func() {
			atomic.AddInt32(ptr, 1)
		})
		hosts = append(hosts, host)
	}

	cli, err := client.Default()
	require.NoError(err)
	cli.Upgrade(hosts)

	for i := 0; i < calls; i++ {
		err = cli.Invoke("test").Do(context.Background())
		require.NoError(err)
	}

	for i := range callCounter {
		value := atomic.LoadInt32(&callCounter[i])
		require.Greater(value, callsPerHost-delta)
		require.Less(value, callsPerHost+delta)
	}
}

func prepareServer(t *testing.T, require *require.Assertions, endpointName string, handler interface{}) string {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(err)
	srv := grpc.NewServer()
	t.Cleanup(func() {
		srv.Shutdown()
	})
	logger, err := log.New()
	require.NoError(err)
	wrapper := endpoint.DefaultWrapper(logger)
	srv.Upgrade(grpc.NewMux().Handle(endpointName, wrapper.Endpoint(handler)))
	go func() {
		err := srv.Serve(listener)
		require.NoError(err)
	}()
	return listener.Addr().String()
}
