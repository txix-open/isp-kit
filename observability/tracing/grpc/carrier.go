// Package grpc provides utilities for gRPC tracing integration.
package grpc

import (
	"google.golang.org/grpc/metadata"
)

// MetadataCarrier implements the TextMapCarrier interface for gRPC metadata.
type MetadataCarrier metadata.MD

// Get returns the first value associated with the given key.
func (m MetadataCarrier) Get(key string) string {
	value := metadata.MD(m).Get(key)
	if len(value) == 0 {
		return ""
	}
	return value[0]
}

// Set associates a key-value pair in the metadata.
func (m MetadataCarrier) Set(key string, value string) {
	metadata.MD(m).Set(key, value)
}

// Keys returns all keys in the metadata.
func (m MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
