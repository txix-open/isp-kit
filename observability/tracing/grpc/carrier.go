package grpc

import (
	"google.golang.org/grpc/metadata"
)

type MetadataCarrier metadata.MD

func (m MetadataCarrier) Get(key string) string {
	value := metadata.MD(m).Get(key)
	if len(value) == 0 {
		return ""
	}
	return value[0]
}

func (m MetadataCarrier) Set(key string, value string) {
	metadata.MD(m).Set(key, value)
}

func (m MetadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
