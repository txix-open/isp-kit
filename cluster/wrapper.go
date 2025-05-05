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

const (
	emitWithAckTimeout = 1 * time.Second
)

type eventFuture struct {
	responseCh chan []byte
	errorCh    chan error
}

type clientWrapper struct {
	cli             *etp.Client
	errorConnChan   chan []byte
	configErrorChan chan []byte
	ctx             context.Context
	eventFutures    sync.Map // map[eventName]eventFuture
	logger          log.Logger
}

func newClientWrapper(ctx context.Context, cli *etp.Client, logger log.Logger) *clientWrapper {
	w := &clientWrapper{
		cli:    cli,
		ctx:    ctx,
		logger: logger,
	}
	w.errorConnChan = w.eventChan(ErrorConnection)
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

	ctx, cancel := context.WithTimeout(ctx, emitWithAckTimeout)
	defer cancel()

	resp, err := w.cli.EmitWithAck(ctx, event, data)
	if err != nil {
		w.logger.Error(ctx, "error", log.Any("error", err))
		return resp, err
	}

	w.logger.Info(ctx, "event acknowledged", log.ByteString("response", resp))
	return resp, nil
}

func (w *clientWrapper) RegisterEvent(event string, handler func([]byte) error) {
	future := eventFuture{
		errorCh:    make(chan error, 1),
		responseCh: make(chan []byte, 1),
	}
	w.eventFutures.Store(event, future)

	w.on(event, func(data []byte) {
		err := handler(data)
		if err != nil {
			sendNonBlocking(future.errorCh, err)
			w.logger.Error(w.ctx, "handle event",
				log.String("event", event),
				log.String("error", err.Error()),
			)
			return
		}
		sendNonBlocking(future.responseCh, data)
	})
}

func (w *clientWrapper) AwaitEvent(ctx context.Context, event string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	futureVal, ok := w.eventFutures.Load(event)
	if !ok {
		return errors.Errorf("event %s not registered", event)
	}
	future := futureVal.(eventFuture) // nolint:forcetypeassert

	select {
	case <-ctx.Done():
		return ctx.Err()
	case data := <-w.errorConnChan:
		return errors.New(string(data))
	case data := <-w.configErrorChan:
		return errors.New(string(data))
	case err := <-future.errorCh:
		return err
	case <-future.responseCh:
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

func (w *clientWrapper) OnDisconnect(handler etp.DisconnectHandler) {
	w.cli.OnDisconnect(handler)
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

func sendNonBlocking[T any](ch chan T, data T) {
	select {
	case ch <- data:
	default:
	}
}
