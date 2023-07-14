package bootstrap

import (
	"context"
	json2 "encoding/json"
	"fmt"
	stdlog "log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/integration-system/isp-kit/app"
	"github.com/integration-system/isp-kit/cluster"
	"github.com/integration-system/isp-kit/config"
	"github.com/integration-system/isp-kit/healthcheck"
	"github.com/integration-system/isp-kit/infra"
	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/rc"
	"github.com/integration-system/isp-kit/rc/schema"
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
}

func New(moduleVersion string, remoteConfig interface{}, endpoints []cluster.EndpointDescriptor) *Bootstrap {
	isDev := strings.ToLower(os.Getenv("APP_MODE")) == "dev"
	localConfigPath, err := configFilePath(isDev)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "resolve local config path"))
		return nil
	}
	configsOpts := []config.Option{
		config.WithValidator(validator.Default),
		config.WithEnvPrefix(os.Getenv("APP_CONFIG_ENV_PREFIX")),
	}
	if localConfigPath != "" {
		configsOpts = append(configsOpts, config.WithExtraSource(config.NewYamlConfig(localConfigPath)))
	}
	app, err := app.New(
		isDev,
		configsOpts...,
	)
	if err != nil {
		stdlog.Fatal(errors.WithMessage(err, "create app"))
		return nil
	}

	boot, err := bootstrap(isDev, app, moduleVersion, remoteConfig, endpoints)
	if err != nil {
		app.Logger().Fatal(app.Context(), err)
	}

	return boot
}

func bootstrap(isDev bool, application *app.Application, moduleVersion string, remoteConfig interface{}, endpoints []cluster.EndpointDescriptor) (*Bootstrap, error) {
	localConfig := LocalConfig{}
	err := application.Config().Read(&localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "read local config")
	}
	if localConfig.GrpcInnerAddress.Port != localConfig.GrpcOuterAddress.Port {
		return nil, errors.Errorf("grpcInnerAddress.port is not equal grpcOuterAddress.port. potential mistake")
	}

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
		GrpcOuterAddress: cluster.AddressConfiguration{
			IP:   broadcastHost,
			Port: strconv.Itoa(localConfig.GrpcOuterAddress.Port),
		},
		Endpoints: endpoints,
	}

	schema := schema.GenerateConfigSchema(remoteConfig)
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
		application.Logger(),
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
	application.Logger().Info(application.Context(),
		"infra server handlers",
		log.Any("infraServerHandlers", []string{"/internal/metrics", "/internal/metrics/descriptions", "/internal/health"}),
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
	}, nil
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
