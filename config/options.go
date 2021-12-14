package config

type Option func(l *Config)

func WithReadingFromFile(file string) Option {
	return func(l *Config) {
		l.file = file
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(l *Config) {
		l.envPrefix = prefix
	}
}

func WithValidator(validator Validator) Option {
	return func(l *Config) {
		l.validator = validator
	}
}
