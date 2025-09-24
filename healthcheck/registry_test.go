package healthcheck_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/healthcheck"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grmqt"
)

// nolint:bodyclose,noctx
func TestRegistry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	mqTest := grmqt.New(test)
	cli := grmqx.New(test.Logger())
	t.Cleanup(cli.Close)
	err := cli.Upgrade(t.Context(), grmqx.NewConfig(mqTest.ConnectionConfig().Url()))
	require.NoError(err)

	registry := healthcheck.NewRegistry(1 * time.Second)
	registry.Register("rabbitClient", cli)

	s := httptest.NewServer(registry.Handler())
	resp, err := s.Client().Get(s.URL)
	require.NoError(err)
	firstResponse, err := io.ReadAll(resp.Body)
	require.NoError(err)
	require.EqualValues(http.StatusOK, resp.StatusCode)

	resp, err = s.Client().Get(s.URL)
	require.NoError(err)
	secondResponse, err := io.ReadAll(resp.Body)
	require.NoError(err)

	require.True(bytes.Equal(firstResponse, secondResponse))

	time.Sleep(1 * time.Second)

	resp, err = s.Client().Get(s.URL)
	require.NoError(err)
	thirdResponse, err := io.ReadAll(resp.Body)
	require.NoError(err)

	require.False(bytes.Equal(firstResponse, thirdResponse))
}
