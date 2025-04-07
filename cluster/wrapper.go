package cluster

import (
	"context"
	"sync"
	"time"

	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
	"github.com/txix-open/etp/v4/msg"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

type clientWrapper struct {
	cli             *etp.Client
	errorConnChan   chan []byte
	errorChan       chan error
	configErrorChan chan []byte
	ctx             context.Context
	logger          log.Logger
	eventStates     sync.Map // map[eventName]chan struct{}
}

func newClientWrapper(ctx context.Context, cli *etp.Client, logger log.Logger) *clientWrapper {
	w := &clientWrapper{
		cli:    cli,
		ctx:    ctx,
		logger: logger,
	}
	w.errorConnChan = w.eventChan(ErrorConnection)
	w.errorChan = make(chan error, 1)
	w.configErrorChan = w.eventChan(ConfigError)
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

func (c *clientWrapper) RegisterEvent(event string, handler func([]byte) error) {
	state := make(chan struct{}, 1)
	c.eventStates.Store(event, state)

	c.on(event, func(data []byte) {
		err := handler(data)
		if err != nil {
			sendNonBlocking(c.errorChan, err)
			c.logger.Error(c.ctx, "handle event",
				log.String("event", event),
				log.String("error", err.Error()),
			)
			return
		}
		sendNonBlocking(state, struct{}{})
	})
}

func (c *clientWrapper) AwaitEvent(ctx context.Context, event string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	val, ok := c.eventStates.Load(event)
	if !ok {
		return errors.Errorf("event %s not registered", event)
	}
	state := val.(chan struct{})

	select {
	case <-ctx.Done():
		return ctx.Err()
	case data := <-c.errorConnChan:
		return errors.New(string(data))
	case data := <-c.configErrorChan:
		return errors.New(string(data))
	case err := <-c.errorChan:
		return err
	case <-state:
		return nil
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

func sendNonBlocking[T any](ch chan T, data T) {
	select {
	case ch <- data:
	default:
	}
}
