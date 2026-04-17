// Package dbt provides test helpers for database operations using dbx.
// It creates isolated test schemas that are automatically cleaned up after
// each test.
package dbt

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/test"
)

// TestDb wraps dbx.Client and provides database operations for testing.
// It automatically creates and manages an isolated schema for each test.
type TestDb struct {
	*dbx.Client

	must   must
	schema string
}

// New creates a new TestDb instance with an isolated schema for testing.
// The schema is automatically created and will be dropped when the test
// completes. Additional dbx options can be provided to customize the
// database client.
//
// nolint:mnd
func New(t *test.Test, opts ...dbx.Option) *TestDb {
	dbConfig := Config(t)
	dbOpenTimeout := t.Config().Optional().Duration("PG_OPEN_TIMEOUT", 15*time.Second)
	ctx, cancel := context.WithTimeout(t.T().Context(), dbOpenTimeout)
	defer cancel()
	opts = append([]dbx.Option{
		dbx.WithCreateSchema(true),
	}, opts...)
	cli, err := dbx.Open(ctx, dbConfig, opts...)
	t.Assert().NoError(err, "open test db cli, schema: %s", dbConfig.Schema)

	db := &TestDb{
		Client: cli,
		schema: dbConfig.Schema,
		must: must{
			db:     cli.Client,
			assert: t.Assert(),
		},
	}
	t.T().Cleanup(func() {
		err := db.Close()
		t.Assert().NoError(err)
	})
	return db
}

// DB returns the underlying dbx client.
func (db *TestDb) DB() (*dbx.Client, error) {
	return db.Client, nil
}

// Must returns a must helper for performing database operations with
// automatic assertion on errors. This is intended for test code.
func (db *TestDb) Must() must {
	return db.must
}

// Schema returns the name of the test schema.
func (db *TestDb) Schema() string {
	return db.schema
}

// Close drops the test schema and closes the database connection.
func (db *TestDb) Close() error {
	_, err := db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", db.schema))
	if err != nil {
		return errors.WithMessage(err, "drop schema")
	}
	err = db.Client.Close()
	return errors.WithMessage(err, "close db")
}

// must provides database operations that automatically assert no errors.
// Each method panics if an error occurs, making it suitable for test code
// where failures should be immediate.
type must struct {
	db     *db.Client
	assert *require.Assertions
}

// Exec executes a SQL query without arguments and panics on error.
func (m must) Exec(query string, args ...any) sql.Result {
	res, err := m.db.Exec(context.Background(), query, args...)
	m.assert.NoError(err)
	return res
}

// Select selects rows into the provided pointer and panics on error.
func (m must) Select(resultPtr any, query string, args ...any) {
	err := m.db.Select(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

// SelectRow selects a single row into the provided pointer and panics on error.
func (m must) SelectRow(resultPtr any, query string, args ...any) {
	err := m.db.SelectRow(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

// ExecNamed executes a SQL query with named arguments and panics on error.
func (m must) ExecNamed(query string, arg any) sql.Result {
	res, err := m.db.ExecNamed(context.Background(), query, arg)
	m.assert.NoError(err)
	return res
}

// Count returns the count of rows matching the query and panics on error.
func (m must) Count(query string, args ...any) int {
	value := 0
	m.SelectRow(&value, query, args...)
	return value
}

// Config creates a dbx.Config for testing with an isolated schema named
// after the test. Connection parameters can be overridden using environment
// variables: PG_HOST, PG_PORT, PG_DB, PG_USER, PG_PASS.
func Config(t *test.Test) dbx.Config {
	cfg := t.Config()
	schema := fmt.Sprintf("test_%s", t.Id()) // name must start from none digit
	dbConfig := dbx.Config{
		Host:        cfg.Optional().String("PG_HOST", "127.0.0.1"),
		Port:        cfg.Optional().Int("PG_PORT", 5432), // nolint:mnd
		Database:    cfg.Optional().String("PG_DB", "test"),
		Username:    cfg.Optional().String("PG_USER", "test"),
		Password:    cfg.Optional().String("PG_PASS", "test"),
		Schema:      schema,
		MaxOpenConn: runtime.NumCPU(),
	}
	return dbConfig
}
