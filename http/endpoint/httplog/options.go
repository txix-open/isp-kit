package httplog

// Option is a function that configures the logConfig.
type Option func(cfg *logConfig)

// WithContentTypes sets the content types for which request and response bodies will be logged.
// Content types are matched using prefix comparison (e.g., "application/json" matches "application/json; charset=utf-8").
func WithContentTypes(logBodyContentTypes []string) Option {
	return func(cfg *logConfig) {
		cfg.logBodyContentTypes = logBodyContentTypes
	}
}

// WithLogBody is a shortcut to enable both request and response body logging.
func WithLogBody(logBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

// WithLogResponseBody enables or disables response body logging.
func WithLogResponseBody(logResponseBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

// WithLogRequestBody enables or disables request body logging.
func WithLogRequestBody(logRequestBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

// WithCombinedLog enables or disables single combined log entry for request and response.
// When enabled, the middleware logs request and response in a single log entry instead of two separate entries.
func WithCombinedLog(enable bool) Option {
	return func(cfg *logConfig) {
		cfg.combinedLog = enable
	}
}
