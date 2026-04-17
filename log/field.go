package log

import (
	"go.uber.org/zap"
)

// Field is an alias for zap.Field representing a structured log field.
type Field = zap.Field

// String creates a string field.
func String(key string, value string) Field {
	return zap.String(key, value)
}

// Int creates an int field.
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int32 creates an int32 field.
func Int32(key string, value int32) Field {
	return zap.Int32(key, value)
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Bool creates a boolean field.
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// ByteString creates a byte slice field.
func ByteString(key string, value []byte) Field {
	return zap.ByteString(key, value)
}

// Any creates a field from any type using Zap's automatic encoding.
func Any(key string, value any) Field {
	return zap.Any(key, value)
}
