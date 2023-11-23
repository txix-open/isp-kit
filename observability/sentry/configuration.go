package sentry

type Config struct {
	Enable        bool
	Dsn           string
	ModuleName    string
	ModuleVersion string
	Environment   string
	InstanceId    string
	Tags          map[string]string
}
