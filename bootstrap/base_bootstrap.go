package bootstrap

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/healthcheck"
	"github.com/txix-open/isp-kit/infra"
	"github.com/txix-open/isp-kit/infra/pprof"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/file"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/app_metrics"
	"github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/validator"
	"go.uber.org/zap/zapcore"
)

const (
	postShutdownWait = 500 * time.Millisecond

	defaultLogFileMaxSizeMb  = 512
	defaultLogFileMaxBackups = 4
	defaultLogFileCompress   = true

	defaultEnableLogSampling       = false
	defaultMaxLogSamplingPerSecond = 1000
	defaulLogSamplingPassEvery     = 100
)

type BaseBootstrap struct {
	App                 *app.Application
	MetricsRegistry     *metrics.Registry
	InfraServer         *infra.Server
	HealthcheckRegistry *healthcheck.Registry
	BindingAddress      string
	MigrationsDir       string
	ModuleName          string
	SentryHub           sentry.Hub
	TracingProvider     tracing.Provider
}

func (b *BaseBootstrap) Fatal(err error) {
	b.SentryHub.CatchError(b.App.Context(), err, log.FatalLevel)
	b.App.Close()
	time.Sleep(postShutdownWait)
	b.App.Logger().Fatal(context.Background(), err)
}

func initApp(isDev bool) (*app.Application, error) {
	appConfig, err := appConfig(isDev)
	if err != nil {
		return nil, errors.WithMessage(err, "app config")
	}
	app, err := app.NewFromConfig(*appConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "create app")
	}
	return app, nil
}

func bootstrap(
	isDev bool,
	application *app.Application,
	sentryHub sentry.Hub,
	localConfig LocalConfig,
	moduleVersion string,
	instanceId string,
) (*BaseBootstrap, error) {
	bindingAddress := net.JoinHostPort(localConfig.GrpcInnerAddress.IP, strconv.Itoa(localConfig.GrpcInnerAddress.Port))

	migrationsDir, err := migrationsDirPath(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve migrations dir path")
	}

	infraServer, metricsRegistry, healthcheckRegistry := initInfra(application, localConfig)

	tracingProvider := initTracing(
		application.Context(),
		sentryHub,
		localConfig,
		moduleVersion,
		instanceId,
		application.Logger(),
	)
	tracing.DefaultProvider = tracingProvider

	application.AddClosers(
		app.CloserFunc(func() error {
			sentryHub.Flush()
			return nil
		}),
		app.CloserFunc(func() error {
			err := tracingProvider.Shutdown(context.Background())
			if err != nil {
				return errors.WithMessage(err, "shutdown tracing provider")
			}
			return nil
		}),
	)

	return &BaseBootstrap{
		App:                 application,
		BindingAddress:      bindingAddress,
		ModuleName:          localConfig.ModuleName,
		MigrationsDir:       migrationsDir,
		InfraServer:         infraServer,
		MetricsRegistry:     metricsRegistry,
		HealthcheckRegistry: healthcheckRegistry,
		SentryHub:           sentryHub,
		TracingProvider:     tracingProvider,
	}, nil
}

func initInfra(application *app.Application, localConfig LocalConfig) (*infra.Server, *metrics.Registry, *healthcheck.Registry) {
	infraServer := infra.NewServer()
	infraServerPort := localConfig.GrpcInnerAddress.Port + 1
	if localConfig.InfraServerPort != 0 {
		infraServerPort = localConfig.InfraServerPort
	}
	infraServerAddress := fmt.Sprintf(":%d", infraServerPort)
	application.AddRunners(
		app.RunnerFunc(func(ctx context.Context) error {
			application.Logger().Info(ctx, "run infra server", log.String("infraServerAddress", infraServerAddress))
			err := infraServer.ListenAndServe(infraServerAddress)
			if err != nil {
				return errors.WithMessagef(err, "run infra server on %s", infraServerAddress)
			}
			return nil
		}),
	)
	application.AddClosers(app.CloserFunc(func() error {
		return infraServer.Shutdown()
	}))

	metricsReg := metrics.DefaultRegistry
	hcReg := healthcheck.NewRegistry(localConfig.HealthcheckHandlerTimeout)

	infraServer.Handle("/internal/metrics", metricsReg.MetricsHandler())
	infraServer.Handle("/internal/metrics/descriptions", metricsReg.MetricsDescriptionHandler())
	infraServer.Handle("/internal/health", hcReg.Handler())
	pprof.RegisterHandlers("/internal", infraServer)

	application.Logger().Info(application.Context(),
		"infra server handlers",
		log.Any("infraServerHandlers",
			append([]string{
				"/internal/metrics",
				"/internal/metrics/descriptions",
				"/internal/health",
			}, pprof.Endpoints("/internal")...),
		),
	)

	return infraServer, metricsReg, hcReg
}

// nolint:ireturn
func initTracing(
	ctx context.Context,
	hub sentry.Hub,
	cfg LocalConfig,
	version string,
	instanceId string,
	logger log.Logger,
) tracing.Provider {
	tracingCfg := tracing.Config{
		Enable:        cfg.Observability.Tracing.Enable,
		Address:       cfg.Observability.Tracing.Address,
		ModuleName:    cfg.ModuleName,
		ModuleVersion: version,
		Environment:   cfg.Observability.Tracing.Environment,
		InstanceId:    instanceId,
		Attributes:    cfg.Observability.Tracing.Attributes,
	}
	provider, err := tracing.NewProviderFromConfiguration(
		ctx,
		logger,
		tracingCfg,
	)
	if err != nil {
		err = errors.WithMessage(err, "new tracing provider, tracing will be disabled")
		hub.CatchError(ctx, err, log.ErrorLevel)
		logger.Error(ctx, err)
		return tracing.NewNoopProvider()
	}
	return provider
}

func appConfig(isDev bool) (*app.Config, error) {
	localConfigPath, err := configFilePath(isDev)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve local config path")
	}
	configsOpts := []config.Option{
		config.WithValidator(validator.Default),
		config.WithEnvPrefix(os.Getenv("APP_CONFIG_ENV_PREFIX")),
	}
	if localConfigPath != "" {
		configsOpts = append(configsOpts, config.WithExtraSource(config.NewYamlConfig(localConfigPath)))
	}

	logConfigSupplier := app.LoggerConfigSupplier(func(cfg *config.Config) log.Config {
		initialLevel := log.InfoLevel
		if isDev {
			initialLevel = log.DebugLevel
		}

		var outputPaths []string
		logFilePath := cfg.Optional().String("LOGFILE.PATH", "")
		if !isDev && logFilePath != "" {
			fileOutput := file.Output{
				File:       logFilePath,
				MaxSizeMb:  cfg.Optional().Int("LOGFILE.MAXSIZEMB", defaultLogFileMaxSizeMb),
				MaxDays:    0,
				MaxBackups: cfg.Optional().Int("LOGFILE.MAXBACKUPS", defaultLogFileMaxBackups),
				Compress:   cfg.Optional().Bool("LOGFILE.COMPRESS", defaultLogFileCompress),
			}
			outputPaths = append(outputPaths, file.ConfigToUrl(fileOutput).String())
		}

		logCounter := app_metrics.NewLogCounter(metrics.DefaultRegistry)

		var sampling *log.SamplingConfig
		isEnableSampling := cfg.Optional().Bool("LOGS.SAMPLING.ENABLE", defaultEnableLogSampling)
		if !isDev && isEnableSampling {
			sampling = &log.SamplingConfig{
				Initial:    cfg.Optional().Int("LOGS.SAMPLING.MAXPERSECOND", defaultMaxLogSamplingPerSecond),
				Thereafter: cfg.Optional().Int("LOGS.SAMPLING.PASSEVERY", defaulLogSamplingPassEvery),
				Hook:       logCounter.DroppedLogCounter(),
			}
		}

		return log.Config{
			IsInDevMode:  isDev,
			OutputPaths:  outputPaths,
			Sampling:     sampling,
			InitialLevel: initialLevel,
			Hooks: []func(entry zapcore.Entry) error{
				logCounter.SampledLogCounter(),
			},
		}
	})

	return &app.Config{
		LoggerConfigSupplier: logConfigSupplier,
		ConfigOptions:        configsOpts,
	}, nil
}

func configPath(isDev bool, configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	if isDev {
		return "./conf/config.json", nil
	}

	return relativePathFromBin("conf/config.json")
}
