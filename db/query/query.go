// Package query provides a fluent query builder for PostgreSQL using the
// Masterminds/squirrel library with dollar-sign placeholder formatting.
package query

import "github.com/Masterminds/squirrel"

// Query is a fluent query builder that wraps squirrel.StatementBuilderType
// with PostgreSQL-compatible dollar-sign placeholders ($1, $2, etc.).
type Query struct {
	squirrel.StatementBuilderType
}

// New creates a new Query instance configured for PostgreSQL with dollar
// sign placeholders. Use this to build SQL queries with a fluent API.
func New() Query {
	return Query{squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}
