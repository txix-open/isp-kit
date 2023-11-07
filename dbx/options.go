package dbx

import (
	"github.com/integration-system/isp-kit/dbx/migration"
	"github.com/integration-system/isp-kit/log"
	"github.com/jackc/pgx/v5"
)

type Option func(db *Client)

func WithMigrationRunner(migrationDir string, logger log.Logger) Option {
	return func(db *Client) {
		db.migrationRunner = migration.NewRunner(migrationDir, logger)
	}
}

func WithTracer(tracer pgx.QueryTracer) Option {
	return func(db *Client) {
		db.tracer = tracer
	}
}
