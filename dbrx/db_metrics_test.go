package dbrx_test

import (
	"context"
	"testing"

	"gitlab.txix.ru/isp/isp-kit/dbx"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/sql_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/sql_tracing"
	"gitlab.txix.ru/isp/isp-kit/test"
	"gitlab.txix.ru/isp/isp-kit/test/dbt"
)

// to see how it works try to modify dbx.NewMetrics struct's methods
func TestDb_WithMetrics(t *testing.T) {
	test, require := test.New(t)
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
	test, require := test.New(t)
	cli := dbt.New(test)
	ctx := context.Background()
	ctx = sql_metrics.OperationLabelToContext(ctx, "test.label")

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestDb_WithMetrics_WithoutLabel(t *testing.T) {
	test, require := test.New(t)
	cli := dbt.New(test, dbx.WithQueryTracer(sql_metrics.NewTracer(metrics.DefaultRegistry)))
	ctx := context.Background()

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}
