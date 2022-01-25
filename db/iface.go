package db

import (
	"context"
	"database/sql"
)

type DB interface {
	Select(ctx context.Context, ptr interface{}, query string, args ...interface{}) error
	SelectRow(ctx context.Context, ptr interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	ExecNamed(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

type Transactional interface {
	RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) error
}
