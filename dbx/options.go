package dbx

import (
	"github.com/jackc/pgx/v5"
	"github.com/txix-open/isp-kit/dbx/migration"
	"github.com/txix-open/isp-kit/log"
)

// Option is a function that configures a Client.
type Option func(db *Client)

// WithMigrationRunner configures the client to run database migrations from the specified directory.
// The logger is used for migration-related logging.
func WithMigrationRunner(migrationDir string, logger log.Logger) Option {
	return func(db *Client) {
		db.migrationRunner = migration.NewRunner(migration.DialectPostgreSQL, migrationDir, logger)
	}
}

// WithQueryTracer registers one or more pgx QueryTracers for query lifecycle events.
func WithQueryTracer(tracers ...pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTraces = append(db.queryTraces, tracers...)
	}
}

// WithCreateSchema enables automatic creation of the database schema if it does not exist.
// This only applies to non-read-only connections with a custom schema.
func WithCreateSchema(createSchema bool) Option {
	return func(db *Client) {
		db.createSchema = createSchema
	}
}

// WithApplicationName sets the application name connection parameter.
// This is used for identification in database logs and monitoring.
func WithApplicationName(moduleName string) Option {
	return func(db *Client) {
		db.applicationName = moduleName
	}
}
