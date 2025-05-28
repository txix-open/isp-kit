package bootstrap

import (
	"context"
	json2 "encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/infra"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/file"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/app_metrics"
	"github.com/txix-open/isp-kit/validator"
	"go.uber.org/zap/zapcore"
)

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
		return "", errors.WithMessage(err, "net dial udp")
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil // nolint:forcetypeassert
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
		if dep.Path == "github.com/txix-open/isp-kit" {
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
				Thereafter: cfg.Optional().Int("LOGS.SAMPLING.PASSEVERY", defaulLogtSamplingPassEvery),
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
