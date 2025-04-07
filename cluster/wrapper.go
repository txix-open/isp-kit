package cluster

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
	"github.com/txix-open/etp/v4/msg"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

type eventState struct {
	mu    sync.Mutex
	data  []byte
	ready atomic.Bool
}

type clientWrapper struct {
	cli             *etp.Client
	errorChan       chan []byte
	configErrorChan chan []byte
	ctx             context.Context
	logger          log.Logger
	eventStates     sync.Map // map[eventName]*eventState
}

func newClientWrapper(ctx context.Context, cli *etp.Client, logger log.Logger) *clientWrapper {
	w := &clientWrapper{
		cli:    cli,
		ctx:    ctx,
		logger: logger,
	}
	errorChan := w.eventChan(ErrorConnection)
	configErrorChan := w.eventChan(ConfigError)
	w.errorChan = errorChan
	w.configErrorChan = configErrorChan
	cli.OnUnknownEvent(etp.HandlerFunc(func(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
		logger.Error(
			ctx,
			"unexpected event from config service",
			log.String("event", event.Name),
			log.ByteString("data", event.Data),
		)
		return nil
	}))
	return w
}

func (w *clientWrapper) on(event string, handler func(data []byte)) {
	w.cli.On(event, etp.HandlerFunc(func(ctx context.Context, conn *etp.Conn, event msg.Event) []byte {
		w.logger.Info(
			w.ctx,
			"event received",
			log.String("event", event.Name),
			log.ByteString("data", hideSecrets(event.Name, event.Data)),
		)
		handler(event.Data)
		return nil
	}))
}

func (w *clientWrapper) EmitJsonWithAck(ctx context.Context, event string, data any) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal '%s' data", event)
	}

	resp, err := w.EmitWithAck(ctx, event, jsonData)
	if err != nil {
		return nil, errors.WithMessage(err, "emit with ask")
	}
	return resp, nil
}

func (w *clientWrapper) EmitWithAck(ctx context.Context, event string, data []byte) ([]byte, error) {
	ctx = log.ToContext(ctx, log.String("event", event))
	w.logger.Info(
		ctx,
		"send event",
		log.ByteString("data", hideSecrets(event, data)),
	)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	resp, err := w.cli.EmitWithAck(ctx, event, data)
	if err != nil {
		w.logger.Error(ctx, "error", log.Any("error", err))
		return resp, err
	}

	w.logger.Info(ctx, "event acknowledged", log.ByteString("response", resp))
	return resp, err
}

func (w *clientWrapper) eventChan(event string) chan []byte {
	ch := make(chan []byte, 1)
	w.on(event, func(data []byte) {
		select {
		case <-w.ctx.Done():
		case ch <- data:
		}
	})
	return ch
}

func (c *clientWrapper) RegisterEvent(event string, handler func([]byte)) {
	state := &eventState{}
	c.eventStates.Store(event, state)

	c.on(event, func(data []byte) {
		state.mu.Lock()
		defer state.mu.Unlock()

		state.data = data
		state.ready.Store(true)
		handler(data)
	})
}

func (c *clientWrapper) AwaitEvent(ctx context.Context, event string, timeout time.Duration) ([]byte, error) {
	val, ok := c.eventStates.Load(event)
	if !ok {
		return nil, errors.Errorf("event %s not registered", event)
	}
	state := val.(*eventState)

	if state.ready.Load() {
		state.mu.Lock()
		defer state.mu.Unlock()
		return state.data, nil
	}

	t := time.NewTimer(timeout)
	defer t.Stop()

	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case data := <-c.errorChan:
			return nil, errors.New(string(data))
		case data := <-c.configErrorChan:
			return nil, errors.New(string(data))
		case <-t.C:
			return nil, errors.Errorf("timeout waiting for event %s", event)
		case <-tick.C:
			if state.ready.Load() {
				state.mu.Lock()
				defer state.mu.Unlock()
				return state.data, nil
			}
		}
	}
}

func (w *clientWrapper) Await(ctx context.Context, ch chan []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-w.errorChan:
		return nil, errors.New(string(data))
	case data := <-w.configErrorChan:
		return nil, errors.New(string(data))
	case data := <-ch:
		return data, nil
	}
}

func (w *clientWrapper) Ping(ctx context.Context) error {
	return w.cli.Ping(ctx)
}

func (w *clientWrapper) Close() error {
	return w.cli.Close()
}

func (w *clientWrapper) Dial(ctx context.Context, url string) error {
	return w.cli.Dial(ctx, url)
}
