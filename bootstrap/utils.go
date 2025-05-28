package bootstrap

import (
	"net"
	"os"
	"path"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/config"
)

func isOnDevMode() bool {
	return strings.ToLower(os.Getenv("APP_MODE")) == "dev"
}

func isOnOfflineMode() bool {
	return strings.ToLower(os.Getenv("CLUSTER_MODE")) == "offline"
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
	conn, err := net.Dial("udp", target) // nolint:noctx
	if err != nil {
		return "", errors.WithMessage(err, "net dial udp")
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.To4().String(), nil // nolint:forcetypeassert
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

// nolint:ireturn
func localConfig[T any](config *config.Config) (T, error) {
	localConfig := new(T)
	err := config.Read(localConfig)
	if err != nil {
		var zero T
		return zero, errors.WithMessage(err, "read local config")
	}
	return *localConfig, nil
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

func defineInfraServerPort(localConfig LocalConfig) int {
	if localConfig.InfraServerPort != 0 {
		return localConfig.InfraServerPort
	}
	return localConfig.GrpcInnerAddress.Port + 1
}
