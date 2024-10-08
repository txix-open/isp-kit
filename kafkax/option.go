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

func (c ConsumerConfig) GetMaxBatchSizeMb() int {
	if c.MaxBatchSizeMb <= 0 {
		return 64
	}

	return c.MaxBatchSizeMb
}

func (c ConsumerConfig) GetCommitInterval() time.Duration {
	if c.CommitIntervalSec == nil {
		return 1 * time.Second
	}

	return time.Duration(*c.CommitIntervalSec) * time.Second
}

func (p PublisherConfig) GetWriteTimeout() time.Duration {
	if p.WriteTimeoutSec == nil {
		return 10 * time.Second
	}

	return time.Duration(*p.WriteTimeoutSec) * time.Second
}

func (p PublisherConfig) GetRequiredAckLevel() kafka.RequiredAcks {
	if p.RequiredAckLevel <= 1 && p.RequiredAckLevel >= -1 {
		return kafka.RequiredAcks(p.RequiredAckLevel)
	}

	return kafka.RequireNone
}

func (p PublisherConfig) GetMaxMessageSizePerPartition() int64 {
	if p.MaxMsgSizeMbPerPartition <= 0 {
		return 64
	}

	return p.MaxMsgSizeMbPerPartition
}

func (p PublisherConfig) GetBatchSizePerPartition() int {
	if p.BatchSizePerPartition <= 0 {
		return 10
	}

	return p.BatchSizePerPartition
}

func (p PublisherConfig) GetBatchTimeoutPerPartition() time.Duration {
	if p.BatchTimeoutPerPartitionMs == nil {
		return 500 * time.Millisecond
	}

	return time.Duration(*p.BatchTimeoutPerPartitionMs) * time.Millisecond
}
