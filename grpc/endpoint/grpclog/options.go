package grpclog

type Option func(cfg *logConfig)

// Shortcut for logResponseBody and logRequestBody
func WithLogBody(logBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

func WithLogResponseBody(logResponseBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

func WithLogRequestBody(logRequestBody bool) Option {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

// Enables single combined request+response log entry.
func WithCombinedLog(enable bool) Option {
	return func(cfg *logConfig) {
		cfg.combinedLog = enable
	}
}
