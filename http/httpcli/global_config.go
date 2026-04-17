package httpcli

import (
	"net/http"
	"time"
)

const (
	defaultTimeout = 15 * time.Second
)

// GlobalRequestConfig holds default settings that apply to all requests
// created by a Client.
type GlobalRequestConfig struct {
	Timeout   time.Duration
	BaseUrl   string
	BasicAuth *BasicAuth
	Cookies   []*http.Cookie
	Headers   map[string]string
}

// NewGlobalRequestConfig creates a new GlobalRequestConfig with default values.
//
// Default timeout is 15 seconds.
func NewGlobalRequestConfig() *GlobalRequestConfig {
	return &GlobalRequestConfig{
		Timeout: defaultTimeout,
	}
}

// configure applies the global settings to a RequestBuilder.
func (c *GlobalRequestConfig) configure(req *RequestBuilder) {
	req.timeout = c.Timeout
	req.baseUrl = c.BaseUrl
	req.basicAuth = c.BasicAuth
	req.cookies = append(req.cookies, c.Cookies...)
	for name, value := range c.Headers {
		req.Header(name, value)
	}
}
