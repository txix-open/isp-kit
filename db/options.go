package db

import (
	"github.com/jackc/pgx/v5"
)

type Option func(db *Client)

func WithQueryTracer(tracer pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTracer = tracer
	}
}
