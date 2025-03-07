package log_middleware

type Option func(cfg *logConfig)

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
