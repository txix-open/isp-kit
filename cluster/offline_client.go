package cluster

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

// OfflineClient provides configuration loading from a local file instead of a remote
// isp-config-service. It is useful for offline or testing scenarios.
type OfflineClient struct {
	configPath string
}

// NewOfflineClient creates a new OfflineClient that reads configuration from the specified path.
func NewOfflineClient(configPath string) OfflineClient {
	return OfflineClient{
		configPath: configPath,
	}
}

// Run loads the configuration from the local file and applies it using the provided
// event handler. Returns nil if the handler is nil.
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

// Close releases resources associated with the offline client.
func (d OfflineClient) Close() error {
	return nil
}

// receiveConfig reads the configuration file and applies it using the event handler.
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
