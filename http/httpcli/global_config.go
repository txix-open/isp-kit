package httpcli

import (
	"net/http"
	"time"
)

const defaultTimeout = 15 * time.Second

type GlobalRequestConfig struct {
	Timeout   time.Duration
	BaseUrl   string
	BasicAuth *BasicAuth
	Cookies   []*http.Cookie
	Headers   map[string]string
}

func NewGlobalRequestConfig() *GlobalRequestConfig {
	return &GlobalRequestConfig{
		Timeout: defaultTimeout,
	}
}

func (c *GlobalRequestConfig) configure(req *RequestBuilder) {
	req.timeout = c.Timeout
	req.baseUrl = c.BaseUrl
	req.basicAuth = c.BasicAuth
	req.cookies = append(req.cookies, c.Cookies...)
	for name, value := range c.Headers {
		req.Header(name, value)
	}
}
