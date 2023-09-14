package log

import (
	"go.uber.org/zap"
)

type Field = zap.Field

func String(key string, value string) Field {
	return zap.String(key, value)
}

func Int(key string, value int) Field {
	return zap.Int(key, value)
}

func Int32(key string, value int32) Field {
	return zap.Int32(key, value)
}

func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

func ByteString(key string, value []byte) Field {
	return zap.ByteString(key, value)
}

func Any(key string, value any) Field {
	return zap.Any(key, value)
}
