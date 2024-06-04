package stompx

import (
	"context"

	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/stompx/consumer"
)

type LogObserver struct {
	logger log.Logger
}

func NewLogObserver(logger log.Logger) LogObserver {
	return LogObserver{logger: logger}
}

func (l LogObserver) Error(c *consumer.Consumer, err error) {
	l.logger.Error(
		context.Background(),
		"stomp client: unexpected consumer error",
		log.String("consumer", c.String()),
		log.Any("error", err),
	)
}

func (l LogObserver) CloseStart(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: closing consumer start",
		log.String("consumer", c.String()),
	)
}

func (l LogObserver) CloseDone(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: closing consumer done",
		log.String("consumer", c.String()),
	)
}

func (l LogObserver) BeginConsuming(c *consumer.Consumer) {
	l.logger.Info(
		context.Background(),
		"stomp client: begin consuming",
		log.String("consumer", c.String()),
	)
}
