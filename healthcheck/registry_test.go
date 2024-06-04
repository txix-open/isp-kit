package healthcheck_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.txix.ru/isp/isp-kit/grmqx"
	"gitlab.txix.ru/isp/isp-kit/healthcheck"
	"gitlab.txix.ru/isp/isp-kit/test"
	"gitlab.txix.ru/isp/isp-kit/test/grmqt"
)

func TestRegistry(t *testing.T) {
	test, require := test.New(t)

	mqTest := grmqt.New(test)
	cli := grmqx.New(test.Logger())
	t.Cleanup(cli.Close)
	err := cli.Upgrade(context.Background(), grmqx.NewConfig(mqTest.ConnectionConfig().Url()))
	require.NoError(err)

	registry := healthcheck.NewRegistry()
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
