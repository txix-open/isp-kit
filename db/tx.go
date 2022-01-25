package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type txOptions struct {
	nativeOpts *sql.TxOptions
}

type TxOption func(options *txOptions)

func IsolationLevel(level sql.IsolationLevel) TxOption {
	return func(options *txOptions) {
		if options.nativeOpts == nil {
			options.nativeOpts = &sql.TxOptions{}
		}
		options.nativeOpts.Isolation = level
	}
}

func ReadOnly() TxOption {
	return func(options *txOptions) {
		if options.nativeOpts == nil {
			options.nativeOpts = &sql.TxOptions{}
		}
		options.nativeOpts.ReadOnly = true
	}
}

type TxFunc func(ctx context.Context, tx *Tx) error

type Tx struct {
	*sqlx.Tx
}

func (t *Tx) Select(ctx context.Context, ptr interface{}, query string, args ...interface{}) error {
	return t.SelectContext(ctx, ptr, query, args...)
}

func (t *Tx) SelectRow(ctx context.Context, ptr interface{}, query string, args ...interface{}) error {
	return t.GetContext(ctx, ptr, query, args...)
}

func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.ExecContext(ctx, query, args...)
}

func (t *Tx) ExecNamed(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return t.NamedExecContext(ctx, query, arg)
}
