package sentry

type Config struct {
	Enable        bool
	Dsn           string
	ModuleName    string
	Environment   string
	ModuleVersion string
}
