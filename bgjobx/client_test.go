package bgjobx_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx"
	"github.com/txix-open/isp-kit/bgjobx/handler"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
)

func TestClient(t *testing.T) {
	t.Parallel()

	test, assert := test.New(t)
	testDb := dbt.New(test, dbx.WithMigrationRunner("./migration", test.Logger()))
	cli := bgjobx.NewClient(testDb, test.Logger())
	t.Cleanup(func() {
		cli.Close()
	})
	callCount := int32(0)
	worker := bgjobx.WorkerConfig{
		Queue: "test",
		Handle: handler.SyncHandlerAdapterFunc(func(ctx context.Context, job bgjob.Job) handler.Result {
			atomic.AddInt32(&callCount, 1)
			return handler.Complete()
		}),
		PollInterval: 500 * time.Millisecond,
	}
	err := cli.Upgrade(t.Context(), []bgjobx.WorkerConfig{worker})
	assert.NoError(err)

	err = cli.Enqueue(t.Context(), bgjob.EnqueueRequest{
		Queue: "test",
		Type:  "type",
	})
	assert.NoError(err)

	time.Sleep(1 * time.Second)
	assert.EqualValues(1, atomic.LoadInt32(&callCount))
}
