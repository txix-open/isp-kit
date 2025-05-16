package dbx

import (
	"github.com/jackc/pgx/v5"
	"github.com/txix-open/isp-kit/dbx/migration"
	"github.com/txix-open/isp-kit/log"
)

type Option func(db *Client)

func WithMigrationRunner(migrationDir string, logger log.Logger) Option {
	return func(db *Client) {
		db.migrationRunner = migration.NewRunner(migration.DialectPostgreSQL, migrationDir, logger)
	}
}

func WithQueryTracer(tracers ...pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTraces = append(db.queryTraces, tracers...)
	}
}

func WithCreateSchema(createSchema bool) Option {
	return func(db *Client) {
		db.createSchema = createSchema
	}
}
