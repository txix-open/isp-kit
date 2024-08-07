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

	return plain.Mechanism{
		Username: auth.Username,
		Password: auth.Password,
	}
}

func (c ConsumerConfig) WithMaxBatchSize() int {
	if c.MaxBatchSizeMb <= 0 {
		return 64
	}

	return c.MaxBatchSizeMb
}

func (c ConsumerConfig) WithCommitIntervalSec() time.Duration {
	if c.CommitIntervalSec == nil {
		return 1 * time.Second
	}

	return time.Duration(*c.CommitIntervalSec) * time.Second
}

func (p PublisherConfig) WithWriteTimeoutSecs() time.Duration {
	if p.WriteTimeoutSec == nil {
		return 10 * time.Second
	}

	return time.Duration(*p.WriteTimeoutSec) * time.Second
}

func (p PublisherConfig) WithRequiredAckLevel() kafka.RequiredAcks {
	if p.RequiredAckLevel <= 1 && p.RequiredAckLevel >= -1 {
		return kafka.RequiredAcks(p.RequiredAckLevel)
	}

	return kafka.RequireNone
}

func (p PublisherConfig) WithMaxMessageSize() int64 {
	if p.MaxMsgSizeMb <= 0 {
		return 1
	}

	return p.MaxMsgSizeMb
}

func (p PublisherConfig) WithBatchSize() int {
	if p.BatchSize <= 0 {
		return 10
	}

	return p.BatchSize
}

func (p PublisherConfig) WithBatchTimeoutMs() time.Duration {
	if p.BatchTimeoutMs == nil {
		return 500 * time.Millisecond
	}

	return time.Duration(*p.BatchTimeoutMs) * time.Millisecond
}
