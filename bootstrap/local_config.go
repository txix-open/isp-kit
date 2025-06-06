package bootstrap

type LocalConfig struct {
	ConfigServiceAddress    ConfigServiceAddr
	GrpcOuterAddress        GrpcOuterAddr
	GrpcInnerAddress        GrpcInnerAddr
	ModuleName              string `validate:"required"`
	DefaultRemoteConfigPath string
	MigrationsDirPath       string
	RemoteConfigOverride    string
	LogFile                 LogFile
	Logs                    Logs
	Observability           Observability
	InfraServerPort         int
	MetricsAutodiscovery    MetricsAutodiscovery
}

type Logs struct {
	Sampling struct {
		Enable       bool
		MaxPerSecond int
		PassEvery    int
	}
}

type LogFile struct {
	Path       string
	MaxSizeMb  int
	MaxBackups int
	Compress   bool
}

type ConfigServiceAddr struct {
	IP   string `validate:"required"`
	Port string `validate:"required"`
}

type GrpcOuterAddr struct {
	IP   string
	Port int `validate:"required"`
}

type GrpcInnerAddr struct {
	IP   string `validate:"required"`
	Port int    `validate:"required"`
}

type Observability struct {
	Sentry  Sentry
	Tracing Tracing
}

type Sentry struct {
	Enable      bool
	Dsn         string
	Environment string
	Tags        map[string]string
}

type Tracing struct {
	Enable      bool
	Address     string
	Environment string
	Attributes  map[string]string
}

type MetricsAutodiscovery struct {
	Enable           bool
	Address          string
	AdditionalLabels map[string]string
}
