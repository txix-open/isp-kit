package dbx_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/integration-system/isp-kit/dbx"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
	}
	db, err := dbx.Open(context.Background(), cfg)
	require.NoError(err)
	var time time.Time
	err = db.SelectRow(context.Background(), &time, "select now()")
	require.NoError(err)
}

func envOrDefault(name string, defValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defValue
}
