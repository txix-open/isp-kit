package client_test

import (
	"net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/log"
)

func TestBalancing(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	servers := 3
	calls := 1500
	callsPerHost := int32(500)
	delta := int32(servers)

	hosts := make([]string, 0)
	callCounter := make([]int32, servers)
	for i := range servers {
		ptr := &callCounter[i]
		host := prepareServer(t, require, "test", func() {
			atomic.AddInt32(ptr, 1)
		})
		hosts = append(hosts, host)
	}

	cli, err := client.Default()
	require.NoError(err)
	cli.Upgrade(hosts)

	for range calls {
		err = cli.Invoke("test").Do(t.Context())
		require.NoError(err)
	}

	for i := range callCounter {
		value := atomic.LoadInt32(&callCounter[i])
		require.GreaterOrEqual(value, callsPerHost-delta)
		require.LessOrEqual(value, callsPerHost+delta)
	}
}

func prepareServer(t *testing.T, require *require.Assertions, endpointName string, handler any) string {
	t.Helper()
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
		assert.NoError(t, err)
	}()
	return listener.Addr().String()
}
