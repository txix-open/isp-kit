package db

import (
	"github.com/jackc/pgx/v5"
)

type Option func(db *Client)

func WithQueryTracer(tracers ...pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTracers = append(db.queryTracers, tracers...)
	}
}
