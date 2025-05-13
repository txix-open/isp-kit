package dbx_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/dbx"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	port, err := strconv.Atoi(envOrDefault("PG_PORT", "5432"))
	require.NoError(err)
	cfg := dbx.Config{
		Host:     envOrDefault("PG_HOST", "127.0.0.1"),
		Port:     port,
		Database: envOrDefault("PG_DB", "test"),
		Username: envOrDefault("PG_USER", "test"),
		Password: envOrDefault("PG_PASS", "test"),
		Params: map[string]string{
			"target_session_attrs": "read-write",
		},
	}
	db, err := dbx.Open(t.Context(), cfg)
	require.NoError(err)
	var time time.Time
	err = db.SelectRow(t.Context(), &time, "select now()")
	require.NoError(err)
}

func envOrDefault(name string, defValue string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}
	return defValue
}
