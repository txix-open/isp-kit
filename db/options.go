package db

import (
	"time"

	"github.com/jackc/pgx/v5"
)

// Option is a function that configures a Client.
type Option func(db *Client)

// WithQueryTracer registers one or more pgx QueryTracers for query lifecycle events.
// Multiple tracers are invoked sequentially for each query.
func WithQueryTracer(tracers ...pgx.QueryTracer) Option {
	return func(db *Client) {
		db.queryTracers = append(db.queryTracers, tracers...)
	}
}

func WithMaxOpenConns(maxConns int32) Option {
	return func(db *Client) {
		db.connSettings.maxConns = maxConns
	}
}

func WithMinIdleConns(minIdleConns int32) Option {
	return func(db *Client) {
		db.connSettings.minIdleConns = minIdleConns
	}
}

func WithMaxConnIdleTime(maxIdleTime time.Duration) Option {
	return func(db *Client) {
		db.connSettings.maxConnsIdleTime = maxIdleTime
	}
}
