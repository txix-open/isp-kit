// Package jsonb provides a custom type for handling JSONB data in PostgreSQL.
// It offers a convenient wrapper around []byte for JSONB storage and retrieval.
package jsonb

// Type is a custom type for JSONB data, represented as a byte slice.
// It is used for storing and retrieving JSON data in PostgreSQL JSONB columns.
type Type = []byte
