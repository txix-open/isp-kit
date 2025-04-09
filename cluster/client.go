package cluster

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/txix-open/etp/v4"
	"github.com/txix-open/isp-kit/json"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/lb"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const (
	livenessProbeFrequency = 5 * time.Second
	livenessProbeTimeout   = 3 * time.Second
	etpClientReadLimit     = 4 * 1024 * 1024
)

type Client struct {
	moduleInfo      ModuleInfo
	configData      ConfigData
	lb              *lb.RoundRobin
	eventHandler    *EventHandler
	logger          log.Logger
	sessionIsActive *atomic.Bool
	closed          *atomic.Bool

	cli *atomic.Pointer[clientWrapper]
}

func NewClient(moduleInfo ModuleInfo, configData ConfigData, hosts []string, logger log.Logger) *Client {
	return &Client{
		moduleInfo:      moduleInfo,
		configData:      configData,
		lb:              lb.NewRoundRobin(hosts),
		sessionIsActive: &atomic.Bool{},
		closed:          &atomic.Bool{},
		cli:             &atomic.Pointer[clientWrapper]{},
		logger:          logger,
	}
}

func (c *Client) Run(ctx context.Context, eventHandler *EventHandler) error {
	c.eventHandler = eventHandler
	c.closed.Store(false)

	for {
		if c.closed.Load() {
			return nil
		}

		host, err := c.lb.Next()
		if err != nil {
			return errors.WithMessage(err, "peek config service host")
		}

		sessionId := requestid.Next()
		ctx = log.ToContext(ctx, log.String("sessionId", sessionId), log.String("configService", host))
		c.logger.Info(ctx, "run session")

		err = c.runSession(ctx, host)
		if errors.Is(err, context.Canceled) {
			return nil
		}
		cli := c.cli.Load()
		if cli != nil {
			cli.Close()
		}
		if err != nil && !etp.IsNormalClose(err) {
			c.logger.Error(ctx, "run config service session", log.String("error", err.Error()))
		}
		c.logger.Info(ctx, "session stopped")

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
		}
	}
}

func (c *Client) Close() error {
	c.closed.Store(true)
	cli := c.cli.Load()
	if cli != nil {
		return cli.Close()
	}
	return nil
}

func (c *Client) Healthcheck(ctx context.Context) error {
	if c.sessionIsActive.Load() {
		return nil
	}
	return errors.New("session inactive")
}

func (c *Client) runSession(ctx context.Context, host string) error {
	defer c.sessionIsActive.Store(false)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	requiredModules := make([]string, 0)
	for moduleName := range c.eventHandler.requiredModules {
		requiredModules = append(requiredModules, moduleName)
	}
	requirements := ModuleRequirements{
		RequiredModules: requiredModules,
		RequireRoutes:   c.eventHandler.routesReceiver != nil,
	}

	etpCli := etp.NewClient(etp.WithClientReadLimit(etpClientReadLimit))
	cli := newClientWrapper(ctx, etpCli, c.logger)
	c.cli.Store(cli)

	disconnectCh := c.subscribeToEvents(cli)

	err := c.dialClientWrapper(ctx, cli, host)
	if err != nil {
		return errors.WithMessage(err, "dial client wrapper")
	}

	_, err = cli.EmitJsonWithAck(ctx, ModuleSendConfigSchema, c.configData)
	if err != nil {
		return errors.WithMessage(err, "emit module config schema")
	}

	_, err = cli.EmitJsonWithAck(ctx, ModuleSendRequirements, requirements)
	if err != nil {
		return errors.WithMessage(err, "emit module requirements")
	}

	err = c.waitModuleReady(ctx, cli, requirements)
	if err != nil {
		return errors.WithMessage(err, "wait module ready")
	}
	err = c.notifyModuleReady(ctx, cli, requirements)
	if err != nil {
		return errors.WithMessage(err, "notify module ready")
	}

	c.sessionIsActive.Store(true)
	go c.livenessProbeLoop(ctx)
	err = <-disconnectCh
	return err
}

func (c *Client) subscribeToEvents(cli *clientWrapper) chan error {
	for moduleName, upgrader := range c.eventHandler.requiredModules {
		event := ModuleConnectedEvent(moduleName)
		cli.RegisterEvent(event, func(data []byte) error {
			hosts, err := readHosts(data)
			if err != nil {
				return errors.WithMessage(err, "read hosts")
			}
			upgrader.Upgrade(hosts)
			return nil
		})
	}

	if c.eventHandler.remoteConfigReceiver != nil {
		cli.RegisterEvent(ConfigSendConfigChanged, c.remoteConfigEventHandler)
		cli.RegisterEvent(ConfigSendConfigWhenConnected, c.remoteConfigEventHandler)
	}
	if c.eventHandler.routesReceiver != nil {
		cli.RegisterEvent(ConfigSendRoutesChanged, c.routesEventHandler)
		cli.RegisterEvent(ConfigSendRoutesWhenConnected, c.routesEventHandler)
	}

	disconnectCh := make(chan error, 1)
	cli.OnDisconnect(func(conn *etp.Conn, err error) {
		disconnectCh <- err
	})

	return disconnectCh
}

func (c *Client) dialClientWrapper(ctx context.Context, cli *clientWrapper, host string) error {
	connUrl, err := url.Parse(fmt.Sprintf("ws://%s/isp-etp/", host))
	if err != nil {
		return errors.WithMessage(err, "parse conn url")
	}
	params := url.Values{}
	params.Add("module_name", c.moduleInfo.ModuleName)
	connUrl.RawQuery = params.Encode()

	err = cli.Dial(ctx, connUrl.String())
	if err != nil {
		return errors.WithMessagef(err, "connect to config service %s", host)
	}
	return nil
}

func (c *Client) remoteConfigEventHandler(data []byte) error {
	cli := c.cli.Load()

	c.logger.Info(cli.ctx, "remote config applying...")
	err := c.applyRemoteConfig(cli.ctx, data)
	if err != nil {
		return errors.WithMessage(err, "apply remote config")
	}
	c.logger.Info(cli.ctx, "remote config successfully applied")
	return nil
}

func (c *Client) routesEventHandler(data []byte) error {
	cli := c.cli.Load()

	routes, err := readRoutes(data)
	if err != nil {
		return errors.WithMessage(err, "read route")
	}
	err = c.eventHandler.routesReceiver.ReceiveRoutes(cli.ctx, routes)
	if err != nil {
		return errors.WithMessage(err, "handle routes")
	}
	return nil
}

func (c *Client) waitModuleReady(ctx context.Context, cli *clientWrapper, requirements ModuleRequirements) error {
	awaitEvents := make(map[string]time.Duration, len(requirements.RequiredModules)+1)
	if requirements.RequireRoutes {
		awaitEvents[ConfigSendRoutesWhenConnected] = time.Second
	}
	if c.eventHandler.remoteConfigReceiver != nil {
		awaitEvents[ConfigSendConfigWhenConnected] = 5*time.Second + c.eventHandler.handleConfigTimeout
	}
	for _, module := range requirements.RequiredModules {
		event := ModuleConnectedEvent(module)
		awaitEvents[event] = time.Second
	}

	for event, timeout := range awaitEvents {
		err := cli.AwaitEvent(ctx, event, timeout)
		if err != nil {
			return errors.WithMessagef(err, "await '%s' event", event)
		}
	}
	return nil
}

func (c *Client) notifyModuleReady(ctx context.Context, cli *clientWrapper, requirements ModuleRequirements) error {
	moduleDependencies := make([]ModuleDependency, 0)
	for _, module := range requirements.RequiredModules {
		dep := ModuleDependency{
			Name:     module,
			Required: true,
		}
		moduleDependencies = append(moduleDependencies, dep)
	}
	declaration := BackendDeclaration{
		ModuleName:      c.moduleInfo.ModuleName,
		Version:         c.moduleInfo.ModuleVersion,
		LibVersion:      c.moduleInfo.LibVersion,
		Endpoints:       c.moduleInfo.Endpoints,
		RequiredModules: moduleDependencies,
		Address:         c.moduleInfo.GrpcOuterAddress,
	}
	_, err := cli.EmitJsonWithAck(ctx, ModuleReady, declaration)
	if err != nil {
		return errors.WithMessage(err, "emit module ready")
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

func (c *Client) livenessProbeLoop(ctx context.Context) {
	for {
		cli := c.cli.Load()
		if cli == nil {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(livenessProbeFrequency):
		}

		ctx, cancel := context.WithTimeout(ctx, livenessProbeTimeout)
		err := cli.Ping(ctx)
		cancel()

		if err != nil {
			c.logger.Warn(ctx, "ping config service", log.String("error", err.Error()))
		}
	}
}

func readRoutes(data []byte) (RoutingConfig, error) {
	var routes RoutingConfig
	err := json.Unmarshal(data, &routes)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal routes")
	}
	return routes, nil
}

func readHosts(data []byte) ([]string, error) {
	addresses := make([]AddressConfiguration, 0)
	err := json.Unmarshal(data, &addresses)
	if err != nil {
		return nil, errors.WithMessagef(err, "unmarshal to address")
	}
	hosts := make([]string, 0)
	for _, addr := range addresses {
		host := net.JoinHostPort(addr.IP, addr.Port)
		hosts = append(hosts, host)
	}
	return hosts, nil
}
