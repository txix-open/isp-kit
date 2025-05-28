package cluster

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

type OfflineClient struct {
	configPath string
}

func NewOfflineClient(configPath string) OfflineClient {
	return OfflineClient{
		configPath: configPath,
	}
}

func (d OfflineClient) Run(ctx context.Context, handler *EventHandler) error {
	if handler == nil {
		return nil
	}

	err := d.receiveConfig(ctx, handler)
	if err != nil {
		return errors.WithMessage(err, "receive config")
	}

	return nil
}

func (d OfflineClient) Close() error {
	return nil
}

func (d OfflineClient) receiveConfig(ctx context.Context, handler *EventHandler) error {
	if handler.remoteConfigReceiver == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, handler.handleConfigTimeout)
	defer cancel()

	cfg, err := os.ReadFile(d.configPath)
	if err != nil {
		return errors.WithMessagef(err, "read file %s", d.configPath)
	}

	err = handler.remoteConfigReceiver.ReceiveConfig(ctx, cfg)
	if err != nil {
		return errors.WithMessage(err, "receive config")
	}
	return nil
}
