package db

import "time"

type ConnectionSettings struct {
	maxConns         int32
	minIdleConns     int32
	maxConnsIdleTime time.Duration
}
