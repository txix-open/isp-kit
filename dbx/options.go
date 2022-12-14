package dbx

import (
	"github.com/jackc/pgx/v5"
)

type Option func(db *Client)

func WithMigration(migrationDir string) Option {
	return func(db *Client) {
		db.withMigration = true
		db.migrationDir = migrationDir
	}
}

func WithTracer(tracer pgx.QueryTracer) Option {
	return func(db *Client) {
		db.tracer = tracer
	}
}
