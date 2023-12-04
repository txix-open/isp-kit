package tracing

type Config struct {
	Enable        bool
	Address       string
	ModuleName    string
	ModuleVersion string
	Environment   string
	InstanceId    string
	Attributes    map[string]string
}
