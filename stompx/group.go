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
	prevCfg   []*consumer.Watcher
	logger    log.Logger
}

func NewConsumerGroup(logger log.Logger) *ConsumerGroup {
	return &ConsumerGroup{
		locker:  &sync.Mutex{},
		prevCfg: []*consumer.Watcher{},
		logger:  logger,
	}
}

func (g *ConsumerGroup) Upgrade(ctx context.Context, consumers ...*consumer.Watcher) error {
	return g.upgrade(ctx, false, consumers...)
}

func (g *ConsumerGroup) UpgradeAndServe(ctx context.Context, consumers ...*consumer.Watcher) {
	_ = g.upgrade(ctx, true, consumers...)
}

func (g *ConsumerGroup) Close() error {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.close()

	return nil
}

func (g *ConsumerGroup) upgrade(ctx context.Context, justServe bool, newConfig ...*consumer.Watcher) error {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.logger.Debug(ctx, "stomp client: received new config")

	if reflect.DeepEqual(g.prevCfg, newConfig) {
		g.logger.Debug(ctx, "stomp client: configs are equal. skipping upgrade")
		return nil
	}

	g.logger.Debug(ctx, "stomp client: initialization began")

	g.close()

	for _, c := range newConfigConsumers(newConfig) {
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

	g.logger.Debug(ctx, "stomp client: initialization done")
	g.prevCfg = newConfig

	return nil
}

func (g *ConsumerGroup) close() {
	for _, c := range g.consumers {
		c.Shutdown()
	}
	g.consumers = nil
}

func newConfigConsumers(cfg []*consumer.Watcher) []*consumer.Watcher {
	consumers := make([]*consumer.Watcher, 0)

	for _, c := range cfg {
		var newConsumer consumer.Watcher
		newConsumer = *c
		consumers = append(consumers, &newConsumer)
	}

	return consumers
}
