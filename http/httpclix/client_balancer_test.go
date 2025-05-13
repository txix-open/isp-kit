package httpclix_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/http/httpclix"
)

func TestClientBalancer_NoHosts(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	err := httpclix.NewClientBalancer(nil).
		Get("/some/endpoint").
		StatusCodeToError().
		DoWithoutResponse(t.Context())
	require.Error(err)
}

func TestClientBalancer(t *testing.T) {
	t.Parallel()
	var (
		require      = require.New(t)
		assert       = assert.New(t)
		hostCount    = 5
		hosts        = make([]string, hostCount)
		callCounters = make([]atomic.Int32, hostCount)
	)
	for i := range hostCount {
		hosts[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.EqualValues("/some/endpoint", r.URL.Path)
			callCounters[i].Add(1)
		})).URL
	}

	var (
		cli       = httpclix.NewClientBalancer(hosts)
		iterCount = 2 * hostCount
	)
	for range iterCount {
		err := cli.Get("/some/endpoint").
			StatusCodeToError().
			DoWithoutResponse(t.Context())
		require.NoError(err)
	}

	var (
		sum       int32
		countList = make([]int32, len(callCounters))
	)
	for i := range callCounters {
		count := callCounters[i].Load()
		require.NotZero(count)
		sum += count
		countList[i] = count
	}
	require.EqualValues(iterCount, sum)

	hosts = hosts[:2]
	cli.Upgrade(hosts)
	for i := range hosts {
		err := cli.Get("/some/endpoint").
			StatusCodeToError().
			DoWithoutResponse(t.Context())
		require.NoError(err)
		countList[i]++
	}

	for i := range callCounters {
		require.EqualValues(countList[i], callCounters[i].Load())
	}
}

func TestClientBalancer_GlobalConfigBaseUrl(t *testing.T) {
	t.Parallel()
	var (
		require     = require.New(t)
		assert      = assert.New(t)
		callCounter atomic.Int32
	)

	baseUrl := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues("/some/endpoint", r.URL.Path)
		callCounter.Add(1)
	})).URL

	hostToIgnore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues("/some/endpoint", r.URL.Path)
		callCounter.Add(-1)
	})).URL

	/* use baseUrl if defined else passed hosts */
	cli := httpclix.NewClientBalancer([]string{hostToIgnore})
	cli.GlobalRequestConfig().BaseUrl = baseUrl

	err := cli.Get("/some/endpoint").
		StatusCodeToError().
		DoWithoutResponse(t.Context())
	require.NoError(err)
	require.EqualValues(1, callCounter.Load())
}
