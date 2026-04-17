package httpcli

// Option is a function that configures a Client.
type Option func(c *Client)

// WithMiddlewares adds one or more middlewares to the client.
// Middlewares are executed in the order they are provided.
func WithMiddlewares(mws ...Middleware) Option {
	return func(c *Client) {
		c.mws = append(c.mws, mws...)
	}
}
