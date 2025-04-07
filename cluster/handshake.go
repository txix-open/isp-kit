package cluster

import (
	"context"
	"fmt"
	"net/url"

	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

type Handshake struct {
	moduleInfo   ModuleInfo
	configData   ConfigData
	requirements ModuleRequirements
	logger       log.Logger
}

func NewHandshake(
	moduleInfo ModuleInfo,
	configData ConfigData,
	requirements ModuleRequirements,
	logger log.Logger,
) *Handshake {
	return &Handshake{
		moduleInfo:   moduleInfo,
		configData:   configData,
		requirements: requirements,
		logger:       logger,
	}
}

func (h Handshake) Do(ctx context.Context, host string) (w *clientWrapper, err error) {
	etpCli := etp.NewClient(etp.WithClientReadLimit(4 * 1024 * 1024))
	cli := newClientWrapper(ctx, etpCli, h.logger)

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

	err = h.emitEvent(ctx, cli, ModuleSendConfigSchema, h.configData)
	if err != nil {
		return nil, errors.WithMessage(err, "emit module config")
	}

	err = h.emitEvent(ctx, cli, ModuleSendRequirements, h.requirements)
	if err != nil {
		return nil, errors.WithMessage(err, "emit module requirements")
	}

	return cli, nil
}

func (h Handshake) EmitModuleReady(ctx context.Context, cli *clientWrapper) error {
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
	err := h.emitEvent(ctx, cli, ModuleReady, declaration)
	if err != nil {
		return errors.WithMessage(err, "emit module ready")
	}
	return nil
}

func (h Handshake) emitEvent(ctx context.Context, cli *clientWrapper, event string, data any) error {
	configData, err := json.Marshal(data)
	if err != nil {
		return errors.WithMessagef(err, "marshal '%s' data", event)
	}

	_, err = cli.EmitWithAck(ctx, event, configData)
	if err != nil {
		return errors.WithMessagef(err, "send '%s' data", event)
	}
	return nil
}
