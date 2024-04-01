package dbrx_test

import (
	"context"
	"testing"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/sql_tracing"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
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
