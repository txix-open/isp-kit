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

type Client struct {
	moduleInfo      ModuleInfo
	configData      ConfigData
	lb              *lb.RoundRobin
	eventHandler    *EventHandler
	logger          log.Logger
	sessionIsActive *atomic.Bool
	closed          *atomic.Bool

	cli *clientWrapper
}

func NewClient(moduleInfo ModuleInfo, configData ConfigData, hosts []string, logger log.Logger) *Client {
	return &Client{
		moduleInfo:      moduleInfo,
		configData:      configData,
		lb:              lb.NewRoundRobin(hosts),
		sessionIsActive: &atomic.Bool{},
		closed:          &atomic.Bool{},
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
	c.closed.Store(true)
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

func (c *Client) Healthcheck(ctx context.Context) error {
	if c.sessionIsActive.Load() {
		return nil
	}
	return errors.New("session inactive")
}

func (c *Client) runSession(ctx context.Context) error {
	defer c.sessionIsActive.Store(false)

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

	etpCli := etp.NewClient(etp.WithClientReadLimit(4 * 1024 * 1024))
	c.cli = newClientWrapper(ctx, etpCli, c.logger)
	c.subscribeToEvents()

	err = c.dialClientWrapper(ctx, host)
	if err != nil {
		return errors.WithMessage(err, "dial client wrapper")
	}
	defer func() {
		if err != nil {
			_ = c.cli.Close()
		}
	}()

	_, err = c.cli.EmitJsonWithAck(ctx, ModuleSendConfigSchema, c.configData)
	if err != nil {
		return errors.WithMessage(err, "emit module config schema")
	}

	_, err = c.cli.EmitJsonWithAck(ctx, ModuleSendRequirements, requirements)
	if err != nil {
		return errors.WithMessage(err, "emit module requirements")
	}

	err = c.waitModuleReady(ctx, requirements)
	if err != nil {
		return errors.WithMessage(err, "wait module ready")
	}
	err = c.notifyModuleReady(ctx, requirements)
	if err != nil {
		return errors.WithMessage(err, "notify module ready")
	}

	c.sessionIsActive.Store(true)
	for {
		err = c.waitAndPing(ctx)
		if err != nil {
			return err
		}
	}
}

func (c *Client) subscribeToEvents() {
	for moduleName, upgrader := range c.eventHandler.requiredModules {
		event := ModuleConnectedEvent(moduleName)
		c.cli.RegisterEvent(event, func(data []byte) error {
			hosts, err := readHosts(data)
			if err != nil {
				return errors.WithMessage(err, "read hosts")
			}
			upgrader.Upgrade(hosts)
			return nil
		})
	}

	c.cli.RegisterEvent(ConfigSendConfigWhenConnected, c.remoteConfigEventHandler)
	c.cli.RegisterEvent(ConfigSendConfigChanged, c.remoteConfigEventHandler)
	c.cli.RegisterEvent(ConfigSendRoutesChanged, c.routesEventHandler)
	c.cli.RegisterEvent(ConfigSendRoutesWhenConnected, c.routesEventHandler)
}

func (c *Client) dialClientWrapper(ctx context.Context, host string) error {
	connUrl, err := url.Parse(fmt.Sprintf("ws://%s/isp-etp/", host))
	if err != nil {
		return errors.WithMessage(err, "parse conn url")
	}
	params := url.Values{}
	params.Add("module_name", c.moduleInfo.ModuleName)
	connUrl.RawQuery = params.Encode()

	err = c.cli.Dial(ctx, connUrl.String())
	if err != nil {
		return errors.WithMessagef(err, "connect to config service %s", host)
	}
	return nil
}

func (c *Client) remoteConfigEventHandler(data []byte) error {
	c.logger.Info(c.cli.ctx, "remote config applying...")
	err := c.applyRemoteConfig(c.cli.ctx, data)
	if err != nil {
		return errors.WithMessage(err, "apply remote config")
	}
	c.logger.Info(c.cli.ctx, "remote config successfully applied")
	return nil
}

func (c *Client) routesEventHandler(data []byte) error {
	routes, err := readRoutes(data)
	if err != nil {
		return errors.WithMessage(err, "read route")
	}
	err = c.eventHandler.routesReceiver.ReceiveRoutes(c.cli.ctx, routes)
	if err != nil {
		return errors.WithMessage(err, "handle routes")
	}
	return nil
}

func (c *Client) waitModuleReady(ctx context.Context, requirements ModuleRequirements) error {
	awaitEvents := make(map[string]time.Duration, len(requirements.RequiredModules)+1)
	awaitEvents[ConfigSendConfigWhenConnected] = 5 * time.Second
	if requirements.RequireRoutes {
		awaitEvents[ConfigSendRoutesWhenConnected] = time.Second
	}
	for _, module := range requirements.RequiredModules {
		event := ModuleConnectedEvent(module)
		awaitEvents[event] = time.Second
	}

	for event, timeout := range awaitEvents {
		err := c.cli.AwaitEvent(ctx, event, timeout)
		if err != nil {
			return errors.WithMessagef(err, "await '%s' event", event)
		}
	}
	return nil
}

func (c *Client) notifyModuleReady(ctx context.Context, requirements ModuleRequirements) error {
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
	_, err := c.cli.EmitJsonWithAck(ctx, ModuleReady, declaration)
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
