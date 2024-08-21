package mcounter

import (
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"golang.org/x/net/context"
	"testing"
	"time"
)

func InitTest(t *testing.T, conf *CounterConfig) (*CounterMetrics, *require.Assertions, context.Context, *dbt.TestDb) {
	ctx := context.Background()
	test, assert := test.New(t)
	testDb := dbt.New(test, dbx.WithMigrationRunner("./migration", test.Logger()))
	rep := NewCounterRepo(testDb)
	txManager := NewTxManager(testDb)
	metricsCli, err := NewCounterMetrics(ctx, metrics.NewRegistry(), test.Logger(), rep, txManager, conf)
	assert.NoError(err)

	return metricsCli, assert, ctx, testDb
}

func TestCloseFlushing(t *testing.T) {
	conf := DefaultConfig().WithFlushInterval(time.Second * 3).WithBufferCap(3)
	metricsCli, assert, ctx, testDb := InitTest(t, conf)

	err := metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	var counters []counter
	err = testDb.Select(ctx, &counters, "SELECT * FROM counter")
	assert.NoError(err)
	assert.Len(counters, 0)

	assert.NoError(metricsCli.Close(ctx))

	err = testDb.Select(ctx, &counters, "SELECT * FROM counter")
	assert.NoError(err)
	assert.Len(counters, 1)
	assert.Equal("test1", counters[0].Name)
}

func TestTimedFlushing(t *testing.T) {
	conf := DefaultConfig().WithFlushInterval(time.Second).WithBufferCap(3)
	metricsCli, assert, ctx, testDb := InitTest(t, conf)

	err := metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	var counters []counter
	err = testDb.Select(ctx, &counters, "SELECT * FROM counter")
	assert.NoError(err)
	assert.Len(counters, 0)

	time.Sleep(time.Second * 3)
	err = testDb.Select(ctx, &counters, "SELECT * FROM counter")
	assert.NoError(err)
	assert.Len(counters, 1)
}

func TestDuplicate(t *testing.T) {
	conf := DefaultConfig().WithFlushInterval(time.Second).WithBufferCap(3)
	metricsCli, assert, ctx, testDb := InitTest(t, conf)

	err := metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)

	assert.NoError(metricsCli.Close(ctx))

	var counterValues []counterValue
	err = testDb.Select(ctx, &counterValues, "SELECT * FROM counter_value where counter_name = 'test1'")
	assert.NoError(err)
	assert.Len(counterValues, 1)
	assert.Equal(3, counterValues[0].AddValue)

	err = metricsCli.Inc("test1", map[string]string{"anotherFieldName": "another value"})
	assert.Equal(DuplicateNameErr, err)
}

func TestBufferOverflow(t *testing.T) {
	conf := DefaultConfig().WithFlushInterval(time.Second * 100).WithBufferCap(1)
	metricsCli, assert, ctx, testDb := InitTest(t, conf)

	err := metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "another value"})
	assert.NoError(err)

	time.Sleep(time.Second)

	var counterValues []counterValue
	err = testDb.Select(ctx, &counterValues, "SELECT * FROM counter_value where counter_name = 'test1'")
	assert.NoError(err)
	assert.Len(counterValues, 2)
}

func TestAddValue(t *testing.T) {
	conf := DefaultConfig().WithFlushInterval(time.Second * 100).WithBufferCap(1)
	metricsCli, assert, ctx, testDb := InitTest(t, conf)

	err := metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "another value"})
	assert.NoError(err)

	time.Sleep(time.Second)

	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "fieldValue1"})
	assert.NoError(err)
	err = metricsCli.Inc("test1", map[string]string{"fieldName1": "another value"})
	assert.NoError(err)

	time.Sleep(time.Second)

	var counterValues []counterValue
	err = testDb.Select(ctx, &counterValues, "SELECT * FROM counter_value where counter_name = 'test1'")
	assert.NoError(err)
	assert.Len(counterValues, 2)

	assert.Equal(2, counterValues[0].AddValue)
}
