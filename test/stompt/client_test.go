package stompt_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-stomp/stomp/v3"
	"gitlab.txix.ru/isp/isp-kit/stompx"
	"gitlab.txix.ru/isp/isp-kit/stompx/publisher"
	"gitlab.txix.ru/isp/isp-kit/test"
	"gitlab.txix.ru/isp/isp-kit/test/stompt"
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
	handler := stompx.NewResultHandler(logger, stompx.AdapterFunc(func(ctx context.Context, msg *stomp.Message) stompx.Result {
		counter.Add(1)
		return stompx.Ack()
	}))
	consumer := stompx.DefaultConsumer(consumerCfg, handler, logger, stompx.ConsumerLog(logger))
	cli.Upgrade(consumer)

	pub := stompx.DefaultPublisher(publisherCfg, stompx.PublisherLog(logger))
	group, ctx := errgroup.WithContext(context.Background())
	group.SetLimit(8)
	for i := 0; i < 100; i++ {
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
