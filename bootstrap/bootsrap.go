package bootstrap

import (
	"context"
	json2 "encoding/json"
	stdlog "log"
	"maps"
	"net"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/observability/sentry"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/validator"
)

type ClusterClient interface {
	Run(ctx context.Context, eventHandler *cluster.EventHandler) error
	Close() error
}

type Bootstrap struct {
	*BaseBootstrap

	ClusterCli   ClusterClient
	RemoteConfig *rc.Config
}

func New(
	moduleVersion string,
	remoteConfig any,
	endpoints []cluster.EndpointDescriptor,
	transport string,
) *Bootstrap {
	isDev := isOnDevMode()
	app, err := initApp(isDev)
	if err != nil {
		stdlog.Fatal(err)
	}

	if transport == cluster.HttpTransport {
		for _, endpoint := range endpoints {
			if endpoint.HttpMethod == "" {
				app.Logger().Fatal(
					app.Context(),
					"invalid endpoint",
					log.String("reason", "empty httpMethod"),
					log.String("path", endpoint.Path),
				)
			}
		}
	}

	localCfg, err := localConfig[ClusteredLocalConfig](app.Config())
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

	if isOnOfflineMode() {
		boot, err := newOfflineClustered(
			isDev,
			app,
			sentryHub,
			localCfg,
			moduleVersion,
		)
		if err != nil {
			err = errors.WithMessage(err, "create offline bootstrap")
			sentryHub.CatchError(app.Context(), err, log.FatalLevel)
			app.Logger().Fatal(app.Context(), err)
		}
		return boot
	}

	boot, err := newClustered(
		isDev,
		app,
		sentryHub,
		localCfg,
		transport,
		remoteConfig,
		moduleVersion,
		endpoints,
	)
	if err != nil {
		err = errors.WithMessage(err, "create bootstrap")
		sentryHub.CatchError(app.Context(), err, log.FatalLevel)
		app.Logger().Fatal(app.Context(), err)
	}
	return boot
}

func newClustered(
	isDev bool,
	application *app.Application,
	sentryHub sentry.Hub,
	localConfig ClusteredLocalConfig,
	transport string,
	remoteConfig any,
	moduleVersion string,
	endpoints []cluster.EndpointDescriptor,
) (*Bootstrap, error) {
	broadcastHost, err := resolveBroadcastHost(localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve broadcast host")
	}

	boot, err := bootstrap(
		isDev,
		application,
		sentryHub,
		localConfig.LocalConfig,
		moduleVersion,
		broadcastHost,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "create bootstrap")
	}

	wrappedLogger := sentry.WrapErrorLogger(application.Logger(), sentryHub)
	clusterCli, err := initClusterClient(
		isDev,
		localConfig,
		remoteConfig,
		moduleVersion,
		transport,
		broadcastHost,
		endpoints,
		wrappedLogger,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "init cluster client")
	}
	boot.HealthcheckRegistry.Register("configServiceConnection", clusterCli)
	rc := rc.New(validator.Default, []byte(localConfig.RemoteConfigOverride))

	return &Bootstrap{
		BaseBootstrap: boot,
		ClusterCli:    clusterCli,
		RemoteConfig:  rc,
	}, nil
}

func newOfflineClustered(
	isDev bool,
	application *app.Application,
	sentryHub sentry.Hub,
	localConfig ClusteredLocalConfig,
	moduleVersion string,
) (*Bootstrap, error) {
	configPath, err := configPath(isDev, localConfig.RemoteConfigPath)
	if err != nil {
		return nil, errors.WithMessage(err, "get config path")
	}

	boot, err := bootstrap(
		isDev,
		application,
		sentryHub,
		localConfig.LocalConfig,
		moduleVersion,
		localConfig.GrpcOuterAddress.IP,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "create bootstrap")
	}

	clusterCli := cluster.NewOfflineClient(configPath)
	rc := rc.New(validator.Default, []byte(localConfig.RemoteConfigOverride))
	return &Bootstrap{
		BaseBootstrap: boot,
		ClusterCli:    clusterCli,
		RemoteConfig:  rc,
	}, nil
}

func resolveBroadcastHost(localConfig ClusteredLocalConfig) (string, error) {
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

func initClusterClient(
	isDev bool,
	localConfig ClusteredLocalConfig,
	remoteConfig any,
	moduleVersion string,
	transport string,
	broadcastHost string,
	endpoints []cluster.EndpointDescriptor,
	logger log.Logger,
) (*cluster.Client, error) {
	moduleInfo := cluster.ModuleInfo{
		ModuleName:    localConfig.ModuleName,
		ModuleVersion: moduleVersion,
		LibVersion:    kitVersion(),
		Transport:     transport,
		GrpcOuterAddress: cluster.AddressConfiguration{
			IP:   broadcastHost,
			Port: strconv.Itoa(localConfig.GrpcOuterAddress.Port),
		},
		Endpoints:            endpoints,
		MetricsAutodiscovery: metricsServiceDiscovery(localConfig, broadcastHost),
	}

	schema := rc.GenerateConfigSchema(remoteConfig)
	schemaData, err := json.Marshal(schema)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal schema")
	}

	path, err := defaultRemoteConfigPath(isDev, localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "resolve path")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "read file")
	}

	defaultConfig := json2.RawMessage{}
	err = json.Unmarshal(data, &defaultConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal json")
	}

	configData := cluster.ConfigData{
		Version: moduleVersion,
		Schema:  schemaData,
		Config:  defaultConfig,
	}

	configServiceHosts, err := parseConfigServiceHosts(localConfig.ConfigServiceAddress)
	if err != nil {
		return nil, errors.WithMessage(err, "parse config service hosts")
	}

	return cluster.NewClient(
		moduleInfo,
		configData,
		configServiceHosts,
		localConfig.RemoteConfigReceiverTimeout,
		logger,
	), nil
}

func defaultRemoteConfigPath(isDev bool, cfg ClusteredLocalConfig) (string, error) {
	if cfg.DefaultRemoteConfigPath != "" {
		return cfg.DefaultRemoteConfigPath, nil
	}

	if isDev {
		return "conf/default_remote_config.json", nil
	}

	return relativePathFromBin("default_remote_config.json")
}

func metricsServiceDiscovery(localConfig ClusteredLocalConfig, broadcastHost string) *cluster.MetricsAutodiscovery {
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
			strconv.Itoa(defineInfraServerPort(localConfig.LocalConfig)),
		)
	}

	return &cluster.MetricsAutodiscovery{
		Address: address,
		Labels:  labels,
	}
}
