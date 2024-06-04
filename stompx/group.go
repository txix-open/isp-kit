package stompx

import (
	"context"
	"sync"

	"gitlab.txix.ru/isp/isp-kit/stompx/consumer"
)

type ConsumerGroup struct {
	locker    sync.Locker
	consumers []*consumer.Watcher
}

func NewConsumerGroup() *ConsumerGroup {
	return &ConsumerGroup{
		locker: &sync.Mutex{},
	}
}

func (g *ConsumerGroup) Upgrade(ctx context.Context, consumers ...consumer.Config) error {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.close()

	for _, config := range consumers {
		c := consumer.NewWatcher(config)
		err := c.Run(ctx)
		if err != nil {
			return err
		}
		g.consumers = append(g.consumers, c)
	}

	return nil
}

func (g *ConsumerGroup) Close() error {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.close()

	return nil
}

func (g *ConsumerGroup) close() {
	for _, c := range g.consumers {
		c.Shutdown()
	}
	g.consumers = nil
}
