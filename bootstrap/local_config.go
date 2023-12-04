package bootstrap

type LocalConfig struct {
	ConfigServiceAddress    ConfigServiceAddr
	GrpcOuterAddress        GrpcOuterAddr
	GrpcInnerAddress        GrpcInnerAddr
	ModuleName              string `valid:"required"`
	DefaultRemoteConfigPath string
	MigrationsDirPath       string
	RemoteConfigOverride    string
	LogFile                 LogFile
	Observability           Observability
	InfraServerPort         int
}

type LogFile struct {
	Path       string
	MaxSizeMb  int
	MaxBackups int
	Compress   bool
}

type ConfigServiceAddr struct {
	IP   string `valid:"required"`
	Port string `valid:"required"`
}

type GrpcOuterAddr struct {
	IP   string
	Port int `valid:"required"`
}

type GrpcInnerAddr struct {
	IP   string `valid:"required"`
	Port int    `valid:"required"`
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
