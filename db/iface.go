package db

import (
	"context"
	"database/sql"
)

type DB interface {
	Select(ctx context.Context, ptr any, query string, args ...any) error
	SelectRow(ctx context.Context, ptr any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error)
}

type Transactional interface {
	RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) error
}
