package bootstrap

import (
	"context"
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

const (
	postShutdownWait = 500 * time.Millisecond

	defaultLogFileMaxSizeMb  = 512
	defaultLogFileMaxBackups = 4
	defaultLogFileCompress   = true

	defaultEnableLogSampling       = false
	defaultMaxLogSamplingPerSecond = 1000
	defaulLogtSamplingPassEvery    = 100
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

func (b *Bootstrap) Fatal(err error) {
	b.SentryHub.CatchError(b.App.Context(), err, log.FatalLevel)
	b.App.Close()
	time.Sleep(postShutdownWait)
	b.App.Logger().Fatal(context.Background(), err)
}

// nolint:funlen
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

	rc := rc.New(validator.Default, []byte(localConfig.RemoteConfigOverride))

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
