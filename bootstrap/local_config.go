package bootstrap

import "time"

// LocalConfig defines the local configuration structure for standalone applications.
//
// Fields:
//   - GrpcOuterAddress: External address configuration (IP and port)
//   - GrpcInnerAddress: Internal address configuration (IP and port, required)
//   - ModuleName: Unique name of the module/application (required)
//   - MigrationsDirPath: Path to database migrations directory (optional)
//   - RemoteConfigOverride: Override content for remote configuration (optional)
//   - LogFile: File-based logging configuration
//   - Logs: Log sampling and rate limiting configuration
//   - Observability: Sentry and tracing configuration
//   - InfraServerPort: Custom port for infrastructure server (optional, defaults to GrpcInnerAddress.Port + 1)
//   - HealthcheckHandlerTimeout: Timeout for health check requests
//   - RemoteConfigPath: Path to application configuration file (optional)
type LocalConfig struct {
	GrpcOuterAddress          GrpcOuterAddr
	GrpcInnerAddress          GrpcInnerAddr
	ModuleName                string `validate:"required"`
	MigrationsDirPath         string
	RemoteConfigOverride      string
	LogFile                   LogFile
	Logs                      Logs
	Observability             Observability
	InfraServerPort           int
	HealthcheckHandlerTimeout time.Duration
	// Path to the application configuration
	RemoteConfigPath string
}

// ClusteredLocalConfig extends LocalConfig with additional configuration for clustered applications.
//
// Embedded:
//   - LocalConfig: All fields from LocalConfig
//
// Additional fields:
//   - ConfigServiceAddress: Address of the config service for cluster coordination (IP;Port format, required)
//   - DefaultRemoteConfigPath: Custom path for default remote configuration file (optional)
//   - RemoteConfigReceiverTimeout: Timeout for receiving remote configuration updates
//   - MetricsAutodiscovery: Configuration for Prometheus metrics auto-discovery
type ClusteredLocalConfig struct {
	LocalConfig

	ConfigServiceAddress        ConfigServiceAddr
	DefaultRemoteConfigPath     string
	RemoteConfigReceiverTimeout time.Duration
	MetricsAutodiscovery        MetricsAutodiscovery
}

// Logs configures log sampling and rate limiting.
//
// Fields:
//   - Sampling.Enable: Enable log sampling to prevent log flooding
//   - Sampling.MaxPerSecond: Maximum log entries per second when sampling is enabled
//   - Sampling.PassEvery: Number of log entries to pass before sampling again
type Logs struct {
	Sampling struct {
		Enable       bool
		MaxPerSecond int
		PassEvery    int
	}
}

// LogFile configures file-based logging output.
//
// Fields:
//   - Path: Path to the log file
//   - MaxSizeMb: Maximum file size in megabytes before rotation (default: 512)
//   - MaxBackups: Maximum number of backup files to keep (default: 4)
//   - Compress: Whether to compress rotated log files (default: true)
type LogFile struct {
	Path       string
	MaxSizeMb  int
	MaxBackups int
	Compress   bool
}

// ConfigServiceAddr defines the address configuration for the cluster config service.
//
// Fields:
//   - IP: Comma-separated list of config service host IPs
//   - Port: Comma-separated list of config service ports (must match length of IP)
//
// Multiple addresses can be specified using semicolon separation (e.g., "host1;host2" and "port1;port2")
type ConfigServiceAddr struct {
	IP   string `validate:"required"`
	Port string `validate:"required"`
}

// GrpcOuterAddr defines the external gRPC address configuration.
//
// Fields:
//   - IP: External IP address (can be empty for standalone)
//   - Port: External gRPC port (required)
type GrpcOuterAddr struct {
	IP   string
	Port int `validate:"required"`
}

// GrpcInnerAddr defines the internal gRPC address configuration.
//
// Fields:
//   - IP: Internal IP address (required)
//   - Port: Internal gRPC port (required)
type GrpcInnerAddr struct {
	IP   string `validate:"required"`
	Port int    `validate:"required"`
}

// Observability configures observability features including Sentry and tracing.
//
// Fields:
//   - Sentry: Error reporting configuration
//   - Tracing: Distributed tracing configuration
type Observability struct {
	Sentry  Sentry
	Tracing Tracing
}

// Sentry configures Sentry error reporting.
//
// Fields:
//   - Enable: Enable Sentry error reporting
//   - Dsn: Sentry DSN (Data Source Name)
//   - Environment: Environment name (e.g., "production", "development")
//   - Tags: Additional tags for Sentry events
type Sentry struct {
	Enable      bool
	Dsn         string
	Environment string
	Tags        map[string]string
}

// Tracing configures distributed tracing (OpenTelemetry).
//
// Fields:
//   - Enable: Enable distributed tracing
//   - Address: Tracing collector endpoint address
//   - Environment: Environment name for trace attribution
//   - Attributes: Additional attributes for trace spans
type Tracing struct {
	Enable      bool
	Address     string
	Environment string
	Attributes  map[string]string
}

// MetricsAutodiscovery configures Prometheus metrics auto-discovery.
//
// Fields:
//   - Enable: Enable metrics auto-discovery
//   - Address: Metrics endpoint address (optional, defaults to infra server address)
//   - AdditionalLabels: Additional labels to attach to metrics
type MetricsAutodiscovery struct {
	Enable           bool
	Address          string
	AdditionalLabels map[string]string
}
