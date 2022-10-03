package httpcli

type Option func(c *Client)

func WithMiddlewares(mws ...Middleware) Option {
	return func(c *Client) {
		c.mws = append(c.mws, mws...)
	}
}

func WithGlobalRequestConfig(cfg GlobalRequestConfig) Option {
	return func(c *Client) {
		c.globalConfig = cfg
	}
}
