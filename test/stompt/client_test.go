package stompt_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/stompx"
	stompConsumer "github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/handler"
	"github.com/txix-open/isp-kit/stompx/publisher"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/stompt"
	"golang.org/x/sync/errgroup"
)

func Test(t *testing.T) {
	t.Parallel()

	test, require := test.New(t)
	logger := test.Logger()
	cli := stompt.New(test)
	publisherCfg := cli.PublisherConfig("test")
	consumerCfg := cli.ConsumerConfig("test")
	consumerCfg.PrefetchCount = 16
	consumerCfg.Concurrency = 16
	counter := &atomic.Int32{}
	handler := stompx.NewResultHandler(logger, handler.AdapterFunc(func(ctx context.Context, msg *stomp.Message) handler.Result {
		counter.Add(1)
		return handler.Ack()
	}))
	consumerConfig := stompx.DefaultConsumer(consumerCfg, handler, logger, stompx.ConsumerLog(logger, true))
	consumer := stompConsumer.NewWatcher(consumerConfig)

	cli.Upgrade(stompx.NewConfig(stompx.WithConsumers(consumer)))

	pub := stompx.DefaultPublisher(publisherCfg, stompx.PublisherLog(logger, true))
	group, ctx := errgroup.WithContext(t.Context())
	group.SetLimit(8)
	for range 100 {
		group.Go(func() error {
			err := pub.Publish(ctx, publisher.PlainText([]byte("hello")))
			return err
		})
	}

	err := group.Wait()
	require.NoError(err)

	time.Sleep(3 * time.Second)
	require.EqualValues(100, counter.Load())
}
