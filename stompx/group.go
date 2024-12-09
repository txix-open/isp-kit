package stompx

import (
	"context"
	"reflect"
	"sync"

	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

type ConsumerGroup struct {
	locker    sync.Locker
	consumers []*consumer.Watcher
	logger    log.Logger
	prevCfg   []consumer.Config
}

func NewConsumerGroup(logger log.Logger) *ConsumerGroup {
	return &ConsumerGroup{
		locker:  &sync.Mutex{},
		logger:  logger,
		prevCfg: []consumer.Config{},
	}
}

func (g *ConsumerGroup) Upgrade(ctx context.Context, consumers ...consumer.Config) error {
	return g.upgrade(ctx, false, consumers...)
}

func (g *ConsumerGroup) UpgradeAndServe(ctx context.Context, consumers ...consumer.Config) {
	_ = g.upgrade(ctx, true, consumers...)
}

func (g *ConsumerGroup) upgrade(ctx context.Context, justServe bool, consumers ...consumer.Config) error {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.logger.Debug(ctx, "stomp client: received new config")

	if reflect.DeepEqual(g.prevCfg, consumers) {
		g.logger.Debug(ctx, "stomp client: configs are equal. skipping upgrade")
		return nil
	}

	g.logger.Debug(ctx, "stomp client: initialization began")

	g.close()

	for _, config := range consumers {
		c := consumer.NewWatcher(config)

		if justServe {
			c.Serve(ctx)
		} else {
			err := c.Run(ctx)
			if err != nil {
				return err
			}
		}

		g.consumers = append(g.consumers, c)
	}

	g.prevCfg = consumers

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
