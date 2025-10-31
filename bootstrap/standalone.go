package bootstrap

import (
	stdlog "log"
	"os"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/validator"
)

type StandaloneBootstrap struct {
	*BaseBootstrap

	appConfigPath string
}

func NewStandalone(moduleVersion string) *StandaloneBootstrap {
	isDev := isOnDevMode()
	app, err := initApp(isDev)
	if err != nil {
		stdlog.Fatal(err)
	}

	localCfg, err := localConfig[LocalConfig](app.Config())
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "create local config"))
	}
	sentryHub, err := sentry.NewHubFromConfiguration(sentry.Config{
		Enable:        localCfg.Observability.Sentry.Enable,
		Dsn:           localCfg.Observability.Sentry.Dsn,
		ModuleName:    localCfg.ModuleName,
		Environment:   localCfg.Observability.Sentry.Environment,
		Tags:          localCfg.Observability.Sentry.Tags,
		InstanceId:    localCfg.GrpcOuterAddress.IP,
		ModuleVersion: moduleVersion,
	})
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "create sentry error reporter"))
	}

	appConfigPath, err := configPath(isDev, localCfg.RemoteConfigPath)
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "get config path"))
	}

	boot, err := bootstrap(
		isDev,
		app,
		sentryHub,
		localCfg,
		moduleVersion,
		localCfg.GrpcOuterAddress.IP,
	)
	if err != nil {
		err = errors.WithMessage(err, "create bootstrap")
		sentryHub.CatchError(app.Context(), err, log.FatalLevel)
		app.Logger().Fatal(app.Context(), err)
	}

	return &StandaloneBootstrap{
		BaseBootstrap: boot,
		appConfigPath: appConfigPath,
	}
}

func (b *StandaloneBootstrap) ReadConfig(destPtr any) error {
	rawCfg, err := os.ReadFile(b.appConfigPath)
	if err != nil {
		return errors.WithMessagef(err, "read file %s", b.appConfigPath)
	}
	err = json.Unmarshal(rawCfg, destPtr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal config")
	}
	err = validator.Default.ValidateToError(destPtr)
	if err != nil {
		return errors.WithMessage(err, "validate config")
	}
	return nil
}
