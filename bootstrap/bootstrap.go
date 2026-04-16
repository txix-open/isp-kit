// Package bootstrap provides a unified initialization framework for isp-kit applications.
//
// The bootstrap package handles application setup including configuration management,
// logging, observability (Sentry and tracing), cluster coordination, and infrastructure
// server initialization. It supports both clustered and standalone deployment modes.
//
// # Usage
//
// For clustered applications:
//
//	boot := bootstrap.New(
//	    "1.0.0",                                    // module version
//	    &RemoteConfig{},                            // remote config struct
//	    []cluster.EndpointDescriptor{},             // service endpoints
//	    cluster.HttpTransport,                      // transport type
//	)
//
// For standalone applications:
//
//	boot := bootstrap.NewStandalone("1.0.0")
//	var cfg MyConfig
//	if err := boot.ReadConfig(&cfg); err != nil {
//	    // handle error
//	}
//
// The package automatically initializes:
//   - Application context and lifecycle management
//   - Logging with optional file output and sampling
//   - Sentry error reporting and tracing
//   - Health check endpoints
//   - Metrics endpoints
//   - Cluster client (for clustered mode)
//   - Remote configuration management (for clustered mode)
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

// ClusterClient defines the interface for cluster coordination operations.
type ClusterClient interface {
	Run(ctx context.Context, eventHandler *cluster.EventHandler) error
	Close() error
}

// Bootstrap represents the main initialization context for clustered applications.
// It embeds BaseBootstrap and adds cluster-specific functionality including
// cluster client management and remote configuration.
//
// The Bootstrap instance provides access to:
//   - Application lifecycle management (via BaseBootstrap)
//   - Cluster coordination through ClusterCli
//   - Remote configuration via RemoteConfig
//
// Create a Bootstrap instance using the New() function.
type Bootstrap struct {
	*BaseBootstrap

	ClusterCli   ClusterClient
	RemoteConfig *rc.Config
}

// New creates and initializes a new Bootstrap instance for clustered applications.
//
// Parameters:
//   - moduleVersion: Version string of the module (e.g., "1.0.0")
//   - remoteConfig: Pointer to a struct defining the remote configuration schema
//   - endpoints: List of service endpoint descriptors for cluster discovery
//   - transport: Transport type (e.g., cluster.HttpTransport or cluster.GrpcTransport)
//
// Returns a fully initialized Bootstrap instance with:
//   - Application context and logging
//   - Sentry error reporting
//   - Cluster client for coordination
//   - Remote configuration management
//   - Infrastructure server with metrics and health endpoints
//
// The function automatically detects the deployment mode (dev/offline) and
// configures the application accordingly. In case of initialization errors,
// the function logs a fatal error and terminates the application.
//
// Example:
//
//	boot := bootstrap.New(
//	    "1.2.3",
//	    &MyRemoteConfig{},
//	    []cluster.EndpointDescriptor{
//	        {Path: "/api/v1", HttpMethod: "GET"},
//	    },
//	    cluster.HttpTransport,
//	)
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
