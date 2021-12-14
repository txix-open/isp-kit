package log

import (
	"go.uber.org/zap"
)

type Field zap.Field

func String(key string, value string) Field {
	return Field(zap.String(key, value))
}

func Int(key string, value int) Field {
	return Field(zap.Int(key, value))
}

func Int32(key string, value int32) Field {
	return Field(zap.Int32(key, value))
}

func Int64(key string, value int64) Field {
	return Field(zap.Int64(key, value))
}

func Bool(key string, value bool) Field {
	return Field(zap.Bool(key, value))
}

func ByteString(key string, value []byte) Field {
	return Field(zap.ByteString(key, value))
}

func Any(key string, value interface{}) Field {
	return Field(zap.Any(key, value))
}
