// Package db provides a PostgreSQL database client wrapper that integrates
// sqlx with pgx, offering transaction support, query tracing, and automatic
// snake_case field mapping for struct tags.
package db

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
)

// Client wraps a sqlx.DB instance with additional functionality for query
// tracing and transaction management. It is safe for concurrent use.
type Client struct {
	*sqlx.DB

	queryTracers tracers
}

// Open establishes a connection to a PostgreSQL database using the provided DSN.
// It applies the given options, validates the connection, and returns a Client.
// Returns an error if the configuration cannot be parsed or the connection fails.
func Open(ctx context.Context, dsn string, opts ...Option) (*Client, error) {
	db := &Client{}
	for _, opt := range opts {
		opt(db)
	}

	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config")
	}
	cfg.Tracer = db.queryTracers

	sqlDb := stdlib.OpenDB(*cfg)

	pgDb := sqlx.NewDb(sqlDb, "pgx")
	pgDb.MapperFunc(ToSnakeCase)
	err = pgDb.PingContext(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "ping database")
	}

	db.DB = pgDb
	return db, nil
}

// RunInTransaction executes the provided function within a database transaction.
// It automatically commits on success and rolls back on error or panic.
// Transaction options can be provided to configure isolation level and other settings.
// Returns an error if the transaction cannot be started, committed, or rolled back.
func (db *Client) RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) (err error) {
	options := &txOptions{}
	for _, opt := range opts {
		opt(options)
	}
	tx, err := db.BeginTxx(
		txContext(ctx, options.metricsLabel),
		options.nativeOpts,
	)
	if err != nil {
		return errors.WithMessage(err, "begin transaction")
	}
	defer func() {
		p := recover()
		if p != nil { // rollback and repanic
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				err = errors.WithMessage(err, rbErr.Error())
			}
			return
		}

		err = tx.Commit()
		if err != nil {
			err = errors.WithMessage(err, "commit tx")
		}
	}()

	return txFunc(ctx, &Tx{tx})
}

// Select executes a query that returns multiple rows and scans them into the provided pointer.
// The pointer must be a slice or a type that sqlx can scan into.
// Returns an error if the query fails or the scan is unsuccessful.
func (db *Client) Select(ctx context.Context, ptr any, query string, args ...any) error {
	return db.SelectContext(ctx, ptr, query, args...)
}

// SelectRow executes a query that returns a single row and scans it into the provided pointer.
// Returns an error if the query fails, no rows are returned, or the scan is unsuccessful.
func (db *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	return db.GetContext(ctx, ptr, query, args...)
}

// Exec executes a query that does not return rows, such as INSERT, UPDATE, or DELETE.
// Returns the sql.Result and any error that occurs during execution.
func (db *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

// ExecNamed executes a named-parameter query that does not return rows.
// The arg parameter should be a struct or map that provides values for named placeholders.
// Returns the sql.Result and any error that occurs during execution.
func (db *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	return db.NamedExecContext(ctx, query, arg)
}

// IsReadOnly checks whether the current database connection is in read-only mode.
// Returns true if the connection is read-only, false otherwise.
// Returns an error if the query fails.
func (db *Client) IsReadOnly(ctx context.Context) (bool, error) {
	var isReadOnly string
	err := db.QueryRowContext(ctx, "SHOW transaction_read_only").Scan(&isReadOnly)
	if err != nil {
		return false, err
	}
	return isReadOnly == "on", nil
}

func txContext(ctx context.Context, label string) context.Context {
	if label == "" {
		return ctx
	}
	return sql_metrics.OperationLabelToContext(ctx, label)
}
