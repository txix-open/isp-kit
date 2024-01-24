package config

type Option func(l *Config)

func WithExtraSource(source Source) Option {
	return func(config *Config) {
		config.extraSources = append(config.extraSources, source)
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(config *Config) {
		config.envPrefix = prefix
	}
}

func WithValidator(validator Validator) Option {
	return func(config *Config) {
		config.validator = validator
	}
}
