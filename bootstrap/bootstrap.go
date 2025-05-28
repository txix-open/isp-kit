package bootstrap

import (
	"context"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/healthcheck"
	"github.com/txix-open/isp-kit/infra"
	"github.com/txix-open/isp-kit/infra/pprof"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/validator"
)

type ClusterClient interface {
	Run(ctx context.Context, eventHandler *cluster.EventHandler) error
	Close() error
	Healthcheck(ctx context.Context) error
}

type Bootstrap struct {
	App                 *app.Application
	ClusterCli          ClusterClient
	RemoteConfig        *rc.Config
	MetricsRegistry     *metrics.Registry
	InfraServer         *infra.Server
	HealthcheckRegistry *healthcheck.Registry
	BindingAddress      string
	MigrationsDir       string
	ModuleName          string
	SentryHub           sentry.Hub
	TracingProvider     tracing.Provider
}

func New(moduleVersion string, remoteConfig any, endpoints []cluster.EndpointDescriptor) *Bootstrap {
	isDev := strings.ToLower(os.Getenv("APP_MODE")) == "dev"
	appConfig, err := appConfig(isDev)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "app config"))
	}
	app, err := app.NewFromConfig(*appConfig)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "create app"))
		return nil
	}

	localConfig, err := localConfig(app.Config())
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "create local config"))
	}

	sentryHub, err := sentry.NewHubFromConfiguration(sentry.Config{
		Enable:        localConfig.Observability.Sentry.Enable,
		Dsn:           localConfig.Observability.Sentry.Dsn,
		ModuleName:    localConfig.ModuleName,
		Environment:   localConfig.Observability.Sentry.Environment,
		Tags:          localConfig.Observability.Sentry.Tags,
		InstanceId:    localConfig.GrpcOuterAddress.IP,
		ModuleVersion: moduleVersion,
	})
	if err != nil {
		app.Logger().Fatal(app.Context(), errors.WithMessage(err, "create sentry error reporter"))
	}

	boot, err := bootstrap(isDev, app, sentryHub, *localConfig, moduleVersion, remoteConfig, endpoints)
	if err != nil {
		err = errors.WithMessage(err, "create bootstrap")
		sentryHub.CatchError(app.Context(), err, log.FatalLevel)
		app.Logger().Fatal(app.Context(), err)
	}

	return boot
}

func (b *Bootstrap) Fatal(err error) {
	b.SentryHub.CatchError(b.App.Context(), err, log.FatalLevel)
	b.App.Close()
	time.Sleep(postShutdownWait)
	b.App.Logger().Fatal(context.Background(), err)
}

func bootstrap(
	isDev bool,
	application *app.Application,
	sentryHub sentry.Hub,
	localConfig LocalConfig,
	moduleVersion string,
	remoteConfig any,
	endpoints []cluster.EndpointDescriptor,
) (*Bootstrap, error) {
	broadcastHost, err := resolveBroadcastHost(localConfig)
	if err != nil {
		return nil, err
	}

	wrappedLogger := sentry.WrapErrorLogger(application.Logger(), sentryHub)
	clusterCli, err := initClusterClient(
		isDev,
		localConfig,
		remoteConfig,
		moduleVersion,
		broadcastHost,
		endpoints,
		wrappedLogger,
	)
	if err != nil {
		return nil, err
	}

	rc := rc.New(validator.Default, []byte(localConfig.RemoteConfigOverride))

	bindingAddress := net.JoinHostPort(localConfig.GrpcInnerAddress.IP, strconv.Itoa(localConfig.GrpcInnerAddress.Port))

	migrationsDir, err := migrationsDirPath(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve migrations dir path")
	}

	infraServer, metricsRegistry, healthcheckRegistry := initInfra(
		application,
		localConfig,
		clusterCli,
	)

	tracingProvider := initTracing(
		application.Context(),
		sentryHub,
		localConfig,
		moduleVersion,
		broadcastHost,
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

	return &Bootstrap{
		App:                 application,
		ClusterCli:          clusterCli,
		RemoteConfig:        rc,
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

func resolveBroadcastHost(localConfig LocalConfig) (string, error) {
	if localConfig.GrpcOuterAddress.IP != "" {
		return localConfig.GrpcOuterAddress.IP, nil
	}
	hosts, err := parseConfigServiceHosts(localConfig.ConfigServiceAddress)
	if err != nil {
		return "", errors.WithMessage(err, "parse config service hosts")
	}
	broadcastHost, err := resolveHost(hosts[0])
	if err != nil {
		return "", errors.WithMessage(err, "resolve local host")
	}
	return broadcastHost, nil
}

func prepareConfigData(isDev bool, localConfig LocalConfig, remoteConfig any, version string) (cluster.ConfigData, error) {
	schema := rc.GenerateConfigSchema(remoteConfig)
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return cluster.ConfigData{}, errors.WithMessage(err, "marshal schema")
	}
	defaultConfig, err := readDefaultRemoteConfig(isDev, localConfig)
	if err != nil {
		return cluster.ConfigData{}, errors.WithMessage(err, "read default remote config")
	}
	return cluster.ConfigData{
		Version: version,
		Schema:  schemaData,
		Config:  defaultConfig,
	}, nil
}

func initClusterClient(
	isDev bool,
	localConfig LocalConfig,
	remoteConfig any,
	moduleVersion string,
	broadcastHost string,
	endpoints []cluster.EndpointDescriptor,
	logger log.Logger,
) (*cluster.Client, error) {
	moduleInfo := cluster.ModuleInfo{
		ModuleName:    localConfig.ModuleName,
		ModuleVersion: moduleVersion,
		LibVersion:    kitVersion(),
		GrpcOuterAddress: cluster.AddressConfiguration{
			IP:   broadcastHost,
			Port: strconv.Itoa(localConfig.GrpcOuterAddress.Port),
		},
		Endpoints:            endpoints,
		MetricsAutodiscovery: metricsServiceDiscovery(localConfig, broadcastHost),
	}

	configData, err := prepareConfigData(isDev, localConfig, remoteConfig, moduleVersion)
	configServiceHosts, err := parseConfigServiceHosts(localConfig.ConfigServiceAddress)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config service hosts")
	}

	return cluster.NewClient(
		moduleInfo,
		configData,
		configServiceHosts,
		logger,
	), nil
}

func initInfra(
	app *app.Application,
	localConfig LocalConfig,
	clusterCli ClusterClient,
) (*infra.Server, *metrics.Registry, *healthcheck.Registry) {
	infra := infraServer(app, localConfig)
	metricsReg := metrics.DefaultRegistry
	hcReg := healthcheck.NewRegistry()

	hcReg.Register("configServiceConnection", clusterCli)

	infra.Handle("/internal/metrics", metricsReg.MetricsHandler())
	infra.Handle("/internal/metrics/descriptions", metricsReg.MetricsDescriptionHandler())
	infra.Handle("/internal/health", hcReg.Handler())
	pprof.RegisterHandlers("/internal", infra)

	app.Logger().Info(app.Context(),
		"infra server handlers",
		log.Any("infraServerHandlers", append([]string{
			"/internal/metrics",
			"/internal/metrics/descriptions",
			"/internal/health",
		}, pprof.Endpoints("/internal")...)),
	)

	return infra, metricsReg, hcReg
}

func infraServer(application *app.Application, localConfig LocalConfig) *infra.Server {
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
	return infraServer
}

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

func defineInfraServerPort(localConfig LocalConfig) int {
	if localConfig.InfraServerPort != 0 {
		return localConfig.InfraServerPort
	}
	return localConfig.GrpcInnerAddress.Port + 1
}

func metricsServiceDiscovery(localConfig LocalConfig, broadcastHost string) *cluster.MetricsAutodiscovery {
	if !localConfig.MetricsAutodiscovery.Enable {
		return nil
	}

	labels := map[string]string{
		"__metrics_path__": "/internal/metrics",
		"app":              localConfig.ModuleName,
	}
	maps.Insert(labels, maps.All(localConfig.MetricsAutodiscovery.AdditionalLabels))

	address := localConfig.MetricsAutodiscovery.Address
	if address == "" {
		address = net.JoinHostPort(
			broadcastHost,
			strconv.Itoa(defineInfraServerPort(localConfig)),
		)
	}

	return &cluster.MetricsAutodiscovery{
		Address: address,
		Labels:  labels,
	}
}
