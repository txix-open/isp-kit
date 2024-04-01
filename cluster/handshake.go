package cluster

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	etpclient "github.com/txix-open/isp-etp-go/v2/client"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

type HandshakeConfirmer interface {
	confirm(ctx context.Context, data HandshakeData) error
}

type Handshake struct {
	moduleInfo   ModuleInfo
	configData   ConfigData
	requirements ModuleRequirements
	confirmer    HandshakeConfirmer
	logger       log.Logger
}

func NewHandshake(
	moduleInfo ModuleInfo,
	configData ConfigData,
	requirements ModuleRequirements,
	confirmer HandshakeConfirmer,
	logger log.Logger,
) *Handshake {
	return &Handshake{
		moduleInfo:   moduleInfo,
		configData:   configData,
		requirements: requirements,
		confirmer:    confirmer,
		logger:       logger,
	}
}

type HandshakeData struct {
	cli                 *clientWrapper
	initialRemoteConfig []byte
	initialRoutes       RoutingConfig
	initialModulesHosts map[string][]string
}

func (h Handshake) Do(ctx context.Context, host string) (w *clientWrapper, err error) {
	etpCli := etpclient.NewClient(etpclient.Config{
		ConnectionReadLimit: 4 * 1024 * 1024,
		HttpClient:          http.DefaultClient,
	})
	cli := newClientWrapper(ctx, etpCli, h.logger)

	remoteConfigChan := cli.EventChan(ConfigSendConfigWhenConnected)
	routesChan := cli.EventChan(ConfigSendRoutesWhenConnected)
	requiredModulesChans := make(map[string]chan []byte)
	for _, module := range h.requirements.RequiredModules {
		event := ModuleConnectedEvent(module)
		requiredModulesChans[module] = cli.EventChan(event)
	}

	connUrl, err := url.Parse(fmt.Sprintf("ws://%s/isp-etp/", host))
	if err != nil {
		return nil, errors.WithMessage(err, "parse conn url")
	}
	params := url.Values{}
	params.Add("module_name", h.moduleInfo.ModuleName)
	connUrl.RawQuery = params.Encode()

	err = cli.Dial(ctx, connUrl.String())
	if err != nil {
		return nil, errors.WithMessagef(err, "connect to config service %s", host)
	}
	defer func() {
		if err != nil {
			_ = cli.Close()
		}
	}()

	configData, err := json.Marshal(h.configData)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal remote config data")
	}

	_, err = cli.EmitWithAck(ctx, ModuleSendConfigSchema, configData)
	if err != nil {
		return nil, errors.WithMessagef(err, "send remote config data")
	}

	requirementsData, err := json.Marshal(h.requirements)
	if err != nil {
		return nil, errors.WithMessagef(err, "marshal module requirements")
	}
	_, err = cli.EmitWithAck(ctx, ModuleSendRequirements, requirementsData)
	if err != nil {
		return nil, errors.WithMessagef(err, "send module requirements")
	}

	remoteConfig, err := cli.Await(ctx, remoteConfigChan, 1*time.Second)
	if err != nil {
		return nil, errors.WithMessage(err, "await remote config")
	}

	var routes RoutingConfig
	if h.requirements.RequireRoutes {
		data, err := cli.Await(ctx, routesChan, 1*time.Second)
		if err != nil {
			return nil, errors.WithMessage(err, "await routes")
		}
		routes, err = readRoutes(data)
		if err != nil {
			return nil, errors.WithMessage(err, "read routes")
		}
	}

	requiredModulesHosts := make(map[string][]string)
	for event, ch := range requiredModulesChans {
		data, err := cli.Await(ctx, ch, 1*time.Second)
		if err != nil {
			return nil, errors.WithMessagef(err, "await event %s", event)
		}
		hosts, err := readHosts(data)
		if err != nil {
			return nil, errors.WithMessagef(err, "read hosts %s", event)
		}
		module := strings.ReplaceAll(event, "_"+ModuleConnectionSuffix, "")
		requiredModulesHosts[module] = hosts
	}

	data := HandshakeData{
		cli:                 cli,
		initialRemoteConfig: remoteConfig,
		initialRoutes:       routes,
		initialModulesHosts: requiredModulesHosts,
	}
	err = h.confirmer.confirm(ctx, data)
	if err != nil {
		return nil, errors.WithMessage(err, "handshake confirm")
	}

	moduleDependencies := make([]ModuleDependency, 0)
	for _, module := range h.requirements.RequiredModules {
		dep := ModuleDependency{
			Name:     module,
			Required: true,
		}
		moduleDependencies = append(moduleDependencies, dep)
	}
	declaration := BackendDeclaration{
		ModuleName:      h.moduleInfo.ModuleName,
		Version:         h.moduleInfo.ModuleVersion,
		LibVersion:      h.moduleInfo.LibVersion,
		Endpoints:       h.moduleInfo.Endpoints,
		RequiredModules: moduleDependencies,
		Address:         h.moduleInfo.GrpcOuterAddress,
	}
	readyData, err := json.Marshal(declaration)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal module ready data")
	}
	_, err = cli.EmitWithAck(ctx, ModuleReady, readyData)
	if err != nil {
		return nil, errors.WithMessage(err, "send module ready data")
	}

	return cli, nil
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
