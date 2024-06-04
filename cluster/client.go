package cluster

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gitlab.txix.ru/isp/isp-kit/lb"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/requestid"
)

type Client struct {
	moduleInfo      ModuleInfo
	configData      ConfigData
	lb              *lb.RoundRobin
	eventHandler    *EventHandler
	logger          log.Logger
	sessionIsActive *atomic.Int32

	cli *clientWrapper
}

func NewClient(moduleInfo ModuleInfo, configData ConfigData, hosts []string, logger log.Logger) *Client {
	return &Client{
		moduleInfo:      moduleInfo,
		configData:      configData,
		lb:              lb.NewRoundRobin(hosts),
		sessionIsActive: &atomic.Int32{},
		logger:          logger,
	}
}

func (c *Client) Run(ctx context.Context, eventHandler *EventHandler) error {
	c.eventHandler = eventHandler

	for {
		err := c.runSession(ctx)
		if errors.Is(err, context.Canceled) {
			return nil
		}

		c.logger.Error(ctx, errors.WithMessage(err, "run config service session"))

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
		}
	}
}

func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

func (c *Client) Healthcheck(ctx context.Context) error {
	if c.sessionIsActive.Load() == int32(1) {
		return nil
	}
	return errors.New("session inactive")
}

func (c *Client) runSession(ctx context.Context) error {
	defer c.sessionIsActive.Store(0)

	host, err := c.lb.Next()
	if err != nil {
		return errors.WithMessage(err, "peek config service host")
	}

	sessionId := requestid.Next()
	ctx = log.ToContext(ctx, log.String("configService", host), log.String("sessionId", sessionId))

	requiredModules := make([]string, 0)
	for moduleName := range c.eventHandler.requiredModules {
		requiredModules = append(requiredModules, moduleName)
	}
	requirements := ModuleRequirements{
		RequiredModules: requiredModules,
		RequireRoutes:   c.eventHandler.routesReceiver != nil,
	}

	handshake := NewHandshake(c.moduleInfo, c.configData, requirements, c, c.logger)
	c.cli, err = handshake.Do(ctx, host)
	if err != nil {
		return errors.WithMessage(err, "do handshake")
	}

	for moduleName := range c.eventHandler.requiredModules {
		event := ModuleConnectedEvent(moduleName)
		upgrader := c.eventHandler.requiredModules[moduleName]
		c.cli.On(event, func(data []byte) {
			hosts, err := readHosts(data)
			if err != nil {
				c.logger.Error(ctx, errors.WithMessage(err, "read hosts"), log.String("event", event))
				return
			}
			upgrader.Upgrade(hosts)
		})
	}
	c.cli.On(ConfigSendConfigChanged, func(data []byte) {
		c.logger.Info(ctx, "remote config applying...")
		err := c.applyRemoteConfig(ctx, data)
		if err != nil {
			c.logger.Error(ctx, errors.WithMessage(err, "apply remote config"), log.String("event", ConfigSendConfigChanged))
			return
		}
		c.logger.Info(ctx, "remote config successfully applied")
	})
	c.cli.On(ConfigSendRoutesChanged, func(data []byte) {
		routes, err := readRoutes(data)
		if err != nil {
			c.logger.Error(ctx, errors.WithMessage(err, "read route"), log.String("event", ConfigSendRoutesChanged))
			return
		}
		err = c.eventHandler.routesReceiver.ReceiveRoutes(ctx, routes)
		if err != nil {
			c.logger.Error(ctx, errors.WithMessage(err, "handle routes"), log.String("event", ConfigSendRoutesChanged))
		}
	})

	c.sessionIsActive.Store(1)

	for {
		err := c.waitAndPing(ctx)
		if err != nil {
			return err
		}
	}
}

func (c *Client) confirm(ctx context.Context, data HandshakeData) error {
	for module, hosts := range data.initialModulesHosts {
		upgrader := c.eventHandler.requiredModules[module]
		upgrader.Upgrade(hosts)
	}

	if c.eventHandler.routesReceiver != nil {
		err := c.eventHandler.routesReceiver.ReceiveRoutes(ctx, data.initialRoutes)
		if err != nil {
			return errors.WithMessagef(err, "receive routes")
		}
	}

	if c.eventHandler.remoteConfigReceiver != nil {
		return c.applyRemoteConfig(ctx, data.initialRemoteConfig)
	}

	return nil
}

func (c *Client) applyRemoteConfig(ctx context.Context, config []byte) (err error) {
	ctx, cancel := context.WithTimeout(ctx, c.eventHandler.handleConfigTimeout)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.eventHandler.remoteConfigReceiver.ReceiveConfig(ctx, config)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) waitAndPing(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(1 * time.Second):
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	err := c.cli.Ping(ctx)
	if err != nil {
		return errors.WithMessage(err, "ping config service")
	}

	return err
}
