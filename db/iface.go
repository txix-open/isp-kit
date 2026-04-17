package db

import (
	"context"
	"database/sql"
)

// DB defines the interface for basic database operations.
type DB interface {
	Select(ctx context.Context, ptr any, query string, args ...any) error
	SelectRow(ctx context.Context, ptr any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error)
}

// Transactional defines the interface for transactional database operations.
type Transactional interface {
	RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) error
}
