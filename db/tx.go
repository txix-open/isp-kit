package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// txOptions holds configuration for database transactions.
type txOptions struct {
	nativeOpts   *sql.TxOptions
	metricsLabel string
}

// TxOption is a function that configures transaction options.
type TxOption func(options *txOptions)

// IsolationLevel sets the transaction isolation level.
func IsolationLevel(level sql.IsolationLevel) TxOption {
	return func(options *txOptions) {
		if options.nativeOpts == nil {
			options.nativeOpts = &sql.TxOptions{}
		}
		options.nativeOpts.Isolation = level
	}
}

// ReadOnly marks the transaction as read-only.
func ReadOnly() TxOption {
	return func(options *txOptions) {
		if options.nativeOpts == nil {
			options.nativeOpts = &sql.TxOptions{}
		}
		options.nativeOpts.ReadOnly = true
	}
}

// MetricsLabel sets a label for transaction metrics.
func MetricsLabel(name string) TxOption {
	return func(options *txOptions) {
		options.metricsLabel = name
	}
}

// TxFunc is a function that executes within a transaction.
type TxFunc func(ctx context.Context, tx *Tx) error

// Tx wraps a sqlx.Tx with convenience methods for database operations.
type Tx struct {
	*sqlx.Tx
}

// Select executes a query that returns multiple rows and scans them into the provided pointer.
func (t *Tx) Select(ctx context.Context, ptr any, query string, args ...any) error {
	return t.SelectContext(ctx, ptr, query, args...)
}

// SelectRow executes a query that returns a single row and scans it into the provided pointer.
func (t *Tx) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	return t.GetContext(ctx, ptr, query, args...)
}

// Exec executes a query that does not return rows.
func (t *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.ExecContext(ctx, query, args...)
}

// ExecNamed executes a named-parameter query that does not return rows.
func (t *Tx) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	return t.NamedExecContext(ctx, query, arg)
}
