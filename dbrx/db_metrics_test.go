package dbrx_test

import (
	"context"
	"github.com/txix-open/isp-kit/dbrx"
	"testing"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/sql_tracing"
	test2 "github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
)

// to see how it works try to modify dbx.NewMetrics struct's methods
func TestDb_WithMetrics(t *testing.T) {
	test, require := test2.New(t)
	option := dbx.WithQueryTracer(
		sql_metrics.NewTracer(metrics.DefaultRegistry),
		dbx.NewLogTracer(test.Logger()),
		sql_tracing.NewConfig().QueryTracer(),
	)
	cli := dbt.New(test, option)
	ctx := context.Background()
	ctx = sql_metrics.OperationLabelToContext(ctx, "test.label")

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestDb_WithoutMetrics(t *testing.T) {
	test, require := test2.New(t)
	cli := dbt.New(test)
	ctx := context.Background()
	ctx = sql_metrics.OperationLabelToContext(ctx, "test.label")

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestDb_WithMetrics_WithoutLabel(t *testing.T) {
	test, require := test2.New(t)
	cli := dbt.New(test, dbx.WithQueryTracer(sql_metrics.NewTracer(metrics.DefaultRegistry)))
	ctx := context.Background()

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestCompareConfig(t *testing.T) {
	test, require := test2.New(t)

	config1 := dbt.Config(test)
	config1.Params = map[string]string{}
	cli := dbrx.New(test.Logger())

	err := cli.Upgrade(context.Background(), config1)
	require.NoError(err)
	db1, err := cli.DB()
	require.NoError(err)

	config2 := dbt.Config(test)
	config2.Params = map[string]string{}
	err = cli.Upgrade(context.Background(), config2)
	require.NoError(err)
	db2, err := cli.DB()
	require.NoError(err)

	require.EqualValues(db1, db2)

	test, require = test2.New(t)
	config3 := dbt.Config(test)
	config3.Params = map[string]string{}
	err = cli.Upgrade(context.Background(), config3)
	require.NoError(err)
	db3, err := cli.DB()
	require.NoError(err)
	require.NotEqualValues(db1, db3)
}
