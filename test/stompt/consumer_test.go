package stompt_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

func TestStompWatcher_Run(t *testing.T) {
	t.Parallel()

	observer := &ErrorCountingObserver{}

	config := consumer.Config{
		Address:          "invalid-address",
		Queue:            "test-queue",
		ReconnectTimeout: 1 * time.Second,
		Observer:         observer,
	}

	watcher := consumer.NewWatcher(config)

	err := watcher.Run(t.Context())

	require.Error(t, err, "expected error due to invalid stomp address")
	require.LessOrEqual(t, observer.ErrorCount.Load(), int32(2), "Number of errors should be <= 2")
}

func TestStompWatcher_Serve(t *testing.T) {
	t.Parallel()

	observer := &ErrorCountingObserver{}

	config := consumer.Config{
		Address:          "invalid-address",
		Queue:            "test-queue",
		ReconnectTimeout: 1 * time.Second,
		Observer:         observer,
	}

	watcher := consumer.NewWatcher(config)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	go watcher.Serve(ctx)

	time.Sleep(3 * time.Second)

	require.GreaterOrEqual(t, observer.ErrorCount.Load(), int32(3), "Number of errors should be >= 3")
}
