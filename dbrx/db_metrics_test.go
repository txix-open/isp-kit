package dbrx_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/integration-system/isp-kit/dbrx"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/stretchr/testify/require"
)

// to see how it works try to modify dbx.NewMetrics struct's methods
func TestDb_WithMetrics(t *testing.T) {
	require := require.New(t)
	cli := prepareCli(require, true)
	ctx := context.Background()
	ctx = dbx.WriteLabelToContext(ctx, "test.label")

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestDb_WithoutMetrics(t *testing.T) {
	require := require.New(t)
	cli := prepareCli(require, false)
	ctx := context.Background()
	ctx = dbx.WriteLabelToContext(ctx, "test.label")

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func TestDb_WithMetrics_WithoutLabel(t *testing.T) {
	require := require.New(t)
	cli := prepareCli(require, true)
	ctx := context.Background()

	var res int
	err := cli.SelectRow(ctx, &res, "select 1")
	require.NoError(err)
	require.Equal(1, res)
}

func prepareCli(require *require.Assertions, withTracer bool) *dbrx.Client {
	var cli *dbrx.Client
	if withTracer {
		cli = dbrx.New(dbx.WithTracer(dbx.NewMetrics()))
	} else {
		cli = dbrx.New()
	}

	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
	}

	err = cli.Upgrade(context.Background(), cfg)
	require.NoError(err)

	return cli
}

func envOrDefault(name string, defValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defValue
}
