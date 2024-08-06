package kafkax

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func PlainAuth(auth *Auth) sasl.Mechanism {
	if auth == nil {
		return plain.Mechanism{}
	}

	if len(auth.Username) == 0 || len(auth.Password) == 0 {
		return plain.Mechanism{}
	}

	return plain.Mechanism{
		Username: auth.Username,
		Password: auth.Password,
	}
}

func WithConsumerMaxBatchSize(batchSize int) int {
	if batchSize <= 0 {
		return 64
	}

	return batchSize
}

func WithCommitIntervalSec(interval int) time.Duration {
	if interval <= 0 {
		return 1 * time.Second
	}

	return time.Duration(interval) * time.Second
}

func WithWriteTimeoutSecs(timeout int) time.Duration {
	return time.Duration(timeout) * time.Second
}

func WithRequiredAckLevel(requireLevel int) kafka.RequiredAcks {
	if requireLevel <= 1 && requireLevel >= -1 {
		return kafka.RequiredAcks(requireLevel)
	}

	return kafka.RequireNone
}

func WithBatchSize(batchSize int) int {
	if batchSize <= 0 {
		return 10
	}

	return batchSize
}

func WithBatchTimeoutMs(timeout int) time.Duration {
	if timeout <= 0 {
		return 500 * time.Millisecond
	}

	return time.Duration(timeout) * time.Millisecond
}
