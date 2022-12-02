package db

import (
	"github.com/jackc/pgx/v5"
)

type Option func(db *Client)

func WithTracer(tracer pgx.QueryTracer) Option {
	return func(db *Client) {
		db.t = tracer
	}
}
