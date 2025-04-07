package cluster

import (
	"context"
	"fmt"
	"net/url"

	"github.com/txix-open/etp/v4"

	"github.com/pkg/errors"
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

	_, err = cli.EmitJsonWithAck(ctx, ModuleSendConfigSchema, h.configData)
	if err != nil {
		return nil, errors.WithMessage(err, "emit module config")
	}

	_, err = cli.EmitJsonWithAck(ctx, ModuleSendRequirements, h.requirements)
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
	_, err := cli.EmitJsonWithAck(ctx, ModuleReady, declaration)
	if err != nil {
		return errors.WithMessage(err, "emit module ready")
	}
	return nil
}
