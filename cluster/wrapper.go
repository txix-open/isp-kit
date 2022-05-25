package cluster

import (
	"context"
	json2 "encoding/json"
	"time"

	etpclient "github.com/integration-system/isp-etp-go/v2/client"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
)

type clientWrapper struct {
	cli             etpclient.Client
	errorChan       chan []byte
	configErrorChan chan []byte
	ctx             context.Context
	logger          log.Logger
}

func newClientWrapper(ctx context.Context, cli etpclient.Client, logger log.Logger) *clientWrapper {
	w := &clientWrapper{
		cli:    cli,
		ctx:    ctx,
		logger: logger,
	}
	errorChan := w.EventChan(ErrorConnection)
	configErrorChan := w.EventChan(ConfigError)
	w.errorChan = errorChan
	w.configErrorChan = configErrorChan
	cli.OnDefault(func(event string, data []byte) {
		logger.Error(ctx, "unexpected event from config service", log.String("event", event), log.Any("data", json2.RawMessage(data)))
	})
	return w
}

func (w *clientWrapper) On(event string, handler func(data []byte)) {
	w.cli.On(event, func(data []byte) {
		copied := make([]byte, len(data))
		copy(copied, data)

		dataToLog := copied
		var err error
		if hidingSecretsEvents[event] {
			dataToLog, err = HideSecrets(copied)
			if err != nil {
				w.logger.Error(w.ctx, "can not hide secrets in config for log", log.Any("error", err))
			}
		}

		w.logger.Info(
			w.ctx,
			"event received",
			log.String("event", event),
			log.Any("data", json2.RawMessage(dataToLog)),
		)
		handler(copied)
	})
}

func (w *clientWrapper) EmitWithAck(ctx context.Context, event string, data []byte) ([]byte, error) {
	ctx = log.ToContext(ctx, log.String("event", event))
	w.logger.Info(
		ctx,
		"send event",
		log.Any("data", json2.RawMessage(data)),
	)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	resp, err := w.cli.EmitWithAck(ctx, event, data)
	if err != nil {
		w.logger.Error(ctx, "error", log.Any("error", err))
		return resp, err
	}

	w.logger.Info(ctx, "event acknowledged", log.String("response", string(resp)))
	return resp, err
}

func (w *clientWrapper) EventChan(event string) chan []byte {
	ch := make(chan []byte, 1)
	w.On(event, func(data []byte) {
		select {
		case <-w.ctx.Done():
		case ch <- data:
		}
	})
	return ch
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

func (w *clientWrapper) IsClosed() bool {
	return w.cli.Closed()
}

func (w *clientWrapper) Dial(ctx context.Context, url string) error {
	return w.cli.Dial(ctx, url)
}
