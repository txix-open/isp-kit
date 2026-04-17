package config

// Option is a function type that configures a Config instance.
// Used with the functional options pattern to customize Config behavior.
type Option func(l *Config)

// WithExtraSource adds a configuration source to be loaded before environment variables.
// Multiple sources can be added and are processed in order.
func WithExtraSource(source Source) Option {
	return func(config *Config) {
		config.extraSources = append(config.extraSources, source)
	}
}

// WithEnvPrefix sets the prefix for environment variable filtering.
// Only environment variables starting with this prefix (case-insensitive) will be loaded.
// The prefix is stripped from the key name when stored.
// Example: WithEnvPrefix("MYAPP") will load MYAPP_HOST as "host".
func WithEnvPrefix(prefix string) Option {
	return func(config *Config) {
		config.envPrefix = prefix
	}
}

// WithValidator sets a Validator for post-decoding configuration validation.
// The validator is called after Read() decodes the configuration into a struct.
func WithValidator(validator Validator) Option {
	return func(config *Config) {
		config.validator = validator
	}
}
