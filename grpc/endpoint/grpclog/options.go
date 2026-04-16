package grpclog

// Option configures logging behavior for server middleware.
type Option func(cfg *logConfig)

// WithLogBody enables logging of both request and response bodies.
func WithLogBody(logBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

// WithLogResponseBody enables or disables logging of response bodies.
func WithLogResponseBody(logResponseBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

// WithLogRequestBody enables or disables logging of request bodies.
func WithLogRequestBody(logRequestBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

// WithCombinedLog enables a single combined log entry for request and response.
// When disabled, requests and responses are logged separately.
func WithCombinedLog(enable bool) Option {
	return func(cfg *logConfig) {
		cfg.combinedLog = enable
	}
}
