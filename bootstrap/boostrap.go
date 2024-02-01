package bootstrap

import (
	"context"
	json2 "encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"path"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/integration-system/isp-kit/app"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/config"
	"github.com/integration-system/isp-kit/healthcheck"
	"github.com/integration-system/isp-kit/infra"
	"github.com/integration-system/isp-kit/infra/pprof"
	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/log/file"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/app_metrics"
	"github.com/integration-system/isp-kit/observability/sentry"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/rc"
	"github.com/integration-system/isp-kit/validator"
	"github.com/pkg/errors"
)

type Bootstrap struct {
	App                 *app.Application
	ClusterCli          *cluster.Client
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

func bootstrap(
	isDev bool,
	application *app.Application,
	sentryHub sentry.Hub,
	localConfig LocalConfig,
	moduleVersion string,
	remoteConfig any,
	endpoints []cluster.EndpointDescriptor,
) (*Bootstrap, error) {
	configServiceHosts, err := parseConfigServiceHosts(localConfig.ConfigServiceAddress)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config service hosts")
	}

	broadcastHost := localConfig.GrpcOuterAddress.IP
	if broadcastHost == "" {
		broadcastHost, err = resolveHost(configServiceHosts[0])
		if err != nil {
			return nil, errors.WithMessage(err, "resolve local host")
		}
	}

	moduleInfo := cluster.ModuleInfo{
		ModuleName:    localConfig.ModuleName,
		ModuleVersion: moduleVersion,
		LibVersion:    kitVersion(),
		GrpcOuterAddress: cluster.AddressConfiguration{
			IP:   broadcastHost,
			Port: strconv.Itoa(localConfig.GrpcOuterAddress.Port),
		},
		Endpoints: endpoints,
	}

	schema := rc.GenerateConfigSchema(remoteConfig)
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal schema")
	}
	defaultConfig, err := readDefaultRemoteConfig(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read default remote config")
	}
	configData := cluster.ConfigData{
		Version: moduleVersion,
		Schema:  schemaData,
		Config:  defaultConfig,
	}

	clusterCli := cluster.NewClient(
		moduleInfo,
		configData,
		configServiceHosts,
		sentry.WrapErrorLogger(application.Logger(), sentryHub),
	)

	delim := localConfig.RemoteConfigOverride.Delim
	if delim == "" {
		delim = "."
	}
	rc := rc.New(validator.Default, []byte(localConfig.RemoteConfigOverride.Data), delim)

	bindingAddress := net.JoinHostPort(localConfig.GrpcInnerAddress.IP, strconv.Itoa(localConfig.GrpcInnerAddress.Port))

	migrationsDir, err := migrationsDirPath(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve migrations dir path")
	}

	healthcheckRegistry := healthcheck.NewRegistry()
	healthcheckRegistry.Register("configServiceConnection", clusterCli)

	infraServer := infraServer(localConfig, application)
	metricsRegistry := metrics.DefaultRegistry
	infraServer.Handle("/internal/metrics", metricsRegistry.MetricsHandler())
	infraServer.Handle("/internal/metrics/descriptions", metricsRegistry.MetricsDescriptionHandler())
	infraServer.Handle("/internal/health", healthcheckRegistry.Handler())
	pprof.RegisterHandlers("/internal", infraServer)
	application.Logger().Info(application.Context(),
		"infra server handlers",
		log.Any("infraServerHandlers", append([]string{
			"/internal/metrics",
			"/internal/metrics/descriptions",
			"/internal/health",
		}, pprof.Endpoints("/internal")...)),
	)

	application.AddClosers(app.CloserFunc(func() error {
		sentryHub.Flush()
		return nil
	}))

	tracingConfig := tracing.Config{
		Enable:        localConfig.Observability.Tracing.Enable,
		Address:       localConfig.Observability.Tracing.Address,
		ModuleName:    localConfig.ModuleName,
		ModuleVersion: moduleVersion,
		Environment:   localConfig.Observability.Tracing.Environment,
		InstanceId:    localConfig.GrpcOuterAddress.IP,
		Attributes:    localConfig.Observability.Tracing.Attributes,
	}
	tracingProvider, err := tracing.NewProviderFromConfiguration(
		application.Context(),
		application.Logger(),
		tracingConfig,
	)
	if err != nil {
		err = errors.WithMessage(err, "new tracing provider, tracing will be disabled")
		sentryHub.CatchError(application.Context(), err, log.ErrorLevel)
		application.Logger().Error(application.Context(), err)
		tracingProvider = tracing.NewNoopProvider()
	}
	tracing.DefaultProvider = tracingProvider
	application.AddClosers(app.CloserFunc(func() error {
		err := tracingProvider.Shutdown(context.Background())
		if err != nil {
			return errors.WithMessage(err, "shutdown tracing provider")
		}
		return nil
	}))

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

func (b *Bootstrap) Fatal(err error) {
	b.SentryHub.CatchError(b.App.Context(), err, log.FatalLevel)
	b.App.Close()
	time.Sleep(500 * time.Millisecond)
	b.App.Logger().Fatal(context.Background(), err)
}

func parseConfigServiceHosts(cfg ConfigServiceAddr) ([]string, error) {
	hosts := strings.Split(cfg.IP, ";")
	ports := strings.Split(cfg.Port, ";")
	if len(hosts) != len(ports) {
		return nil, errors.New("len(hosts) != len(ports)")
	}
	arr := make([]string, 0)
	for i, host := range hosts {
		arr = append(arr, net.JoinHostPort(host, ports[i]))
	}
	return arr, nil
}

func resolveHost(target string) (string, error) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil
}

func readDefaultRemoteConfig(isDev bool, cfg LocalConfig) (json2.RawMessage, error) {
	path, err := defaultRemoteConfigPath(isDev, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "read file")
	}

	remoteConfig := json2.RawMessage{}
	err = json.Unmarshal(data, &remoteConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal json")
	}

	return remoteConfig, nil
}

func defaultRemoteConfigPath(isDev bool, cfg LocalConfig) (string, error) {
	if cfg.DefaultRemoteConfigPath != "" {
		return cfg.DefaultRemoteConfigPath, nil
	}

	if isDev {
		return "conf/default_remote_config.json", nil
	}

	return relativePathFromBin("default_remote_config.json")
}

func configFilePath(isDev bool) (string, error) {
	cfgPath := os.Getenv("APP_CONFIG_PATH")
	if cfgPath != "" {
		return cfgPath, nil
	}

	if isDev {
		return "./conf/config_dev.yml", nil
	}

	return relativePathFromBin("config.yml")
}

func migrationsDirPath(isDev bool, cfg LocalConfig) (string, error) {
	if cfg.MigrationsDirPath != "" {
		return cfg.MigrationsDirPath, nil
	}

	if isDev {
		return "./migrations", nil
	}

	return relativePathFromBin("migrations")
}

func relativePathFromBin(part string) (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", errors.WithMessage(err, "get executable path")
	}
	return path.Join(path.Dir(ex), part), nil
}

func infraServer(localConfig LocalConfig, application *app.Application) *infra.Server {
	infraServer := infra.NewServer()
	infraServerPort := localConfig.GrpcInnerAddress.Port + 1
	if localConfig.InfraServerPort != 0 {
		infraServerPort = localConfig.InfraServerPort
	}
	infraServerAddress := fmt.Sprintf(":%d", infraServerPort)
	application.AddRunners(app.RunnerFunc(func(ctx context.Context) error {
		application.Logger().Info(ctx, "run infra server", log.String("infraServerAddress", infraServerAddress))
		err := infraServer.ListenAndServe(infraServerAddress)
		if err != nil {
			return errors.WithMessagef(err, "run infra server on %s", infraServerAddress)
		}
		return nil
	}))
	application.AddClosers(app.CloserFunc(func() error {
		return infraServer.Shutdown()
	}))
	return infraServer
}

func localConfig(config *config.Config) (*LocalConfig, error) {
	localConfig := LocalConfig{}
	err := config.Read(&localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}
	if localConfig.GrpcInnerAddress.Port != localConfig.GrpcOuterAddress.Port {
		return nil, errors.Errorf("grpcInnerAddress.port is not equal grpcOuterAddress.port. potential mistake")
	}
	return &localConfig, nil
}

func kitVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "0.0.0"
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/integration-system/isp-kit" {
			return dep.Version
		}
	}
	return "0.0.0"
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

		var fileOutput *file.Output
		logFilePath := cfg.Optional().String("LOGFILE.PATH", "")
		if !isDev && logFilePath != "" {
			fileOutput = &file.Output{
				File:       logFilePath,
				MaxSizeMb:  cfg.Optional().Int("LOGFILE.MAXSIZEMB", 512),
				MaxDays:    0,
				MaxBackups: cfg.Optional().Int("LOGFILE.MAXBACKUPS", 4),
				Compress:   cfg.Optional().Bool("LOGFILE.COMPRESS", true),
			}
		}

		var sampling *log.SamplingConfig
		isEnableSampling := cfg.Optional().Bool("LOGS.SAMPLING.ENABLE", true)
		if !isDev && isEnableSampling {
			sampling = &log.SamplingConfig{
				Initial:    cfg.Optional().Int("LOGS.SAMPLING.MAXPERSECOND", 1000),
				Thereafter: cfg.Optional().Int("LOGS.SAMPLING.PASSEVERY", 100),
				Hook:       app_metrics.LogSamplingHook(metrics.DefaultRegistry),
			}
		}

		return log.Config{
			IsInDevMode:  isDev,
			FileOutput:   fileOutput,
			Sampling:     sampling,
			InitialLevel: initialLevel,
		}
	})

	return &app.Config{
		LoggerConfigSupplier: logConfigSupplier,
		ConfigOptions:        configsOpts,
	}, nil
}
