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

func WithWriteTimeoutSecs(timeout int) time.Duration {
	return time.Duration(timeout) * time.Second
}

func WithRequiredAckLevel(requireLevel int) kafka.RequiredAcks {
	if requireLevel <= 1 && requireLevel >= -1 {
		return kafka.RequiredAcks(requireLevel)
	}

	return kafka.RequireNone
}
