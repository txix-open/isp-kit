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

// StandaloneBootstrap represents the initialization context for standalone applications.
// It embeds BaseBootstrap and provides functionality for reading application configuration
// from a local file.
//
// Use this type for applications that do not require cluster coordination or remote
// configuration. Ideal for single-instance services, batch jobs, or development environments.
//
// Create a StandaloneBootstrap using the NewStandalone() function.
type StandaloneBootstrap struct {
	*BaseBootstrap

	appConfigPath string
}

// NewStandalone creates and initializes a new StandaloneBootstrap instance.
//
// Parameters:
//   - moduleVersion: Version string of the module (e.g., "1.0.0")
//
// Returns a fully initialized StandaloneBootstrap with:
//   - Application context and logging
//   - Sentry error reporting
//   - Infrastructure server with metrics and health endpoints
//   - Tracing provider (if configured)
//
// Unlike New(), this function does not initialize cluster client or remote
// configuration. It is suitable for standalone applications that read their
// configuration from local files.
//
// Example:
//
//	boot := bootstrap.NewStandalone("1.2.3")
//	var cfg MyConfig
//	if err := boot.ReadConfig(&cfg); err != nil {
//	    // handle configuration error
//	}
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

// ReadConfig reads and validates the application configuration from a JSON file.
//
// Parameters:
//   - destPtr: Pointer to a struct where the configuration will be unmarshaled
//
// Returns an error if:
//   - The configuration file cannot be read
//   - The JSON cannot be unmarshaled into the destination struct
//   - The configuration fails validation
//
// The configuration file path is determined by the RemoteConfigPath field
// in the local configuration, or defaults to "./conf/config.json" in dev mode
// or "./conf/config.json" relative to the binary in production.
//
// Example:
//
//	var cfg MyConfig
//	if err := boot.ReadConfig(&cfg); err != nil {
//	    log.Fatalf("failed to read config: %v", err)
//	}
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
