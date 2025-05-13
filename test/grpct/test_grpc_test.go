package grpct_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grpct"
)

func TestMockServer_Mock(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	srv, cli := grpct.NewMock(test)
	srv.Mock("endpoint", func() {

	})
	err := cli.Invoke("endpoint").Do(t.Context())
	require.NoError(err)
}
