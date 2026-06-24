package httpcli

import (
	"net/http"
	"time"

	"github.com/txix-open/isp-kit/codec"
)

const (
	defaultTimeout         = 15 * time.Second
	defaultEncodeThreshold = codec.DefaultEncodeThreshold
)

// GlobalRequestConfig holds default settings that apply to all requests
// created by a Client.
type GlobalRequestConfig struct {
	// Timeout specifies the request timeout.
	Timeout time.Duration

	// BaseUrl specifies the base URL for requests.
	BaseUrl string

	// BasicAuth specifies HTTP Basic Authentication credentials.
	BasicAuth *BasicAuth

	// Cookies contains cookies sent with every request.
	Cookies []*http.Cookie

	// Headers contains headers sent with every request.
	Headers map[string]string

	// AcceptEncodedResponse enables automatic acceptance of encoded responses.
	AcceptEncodedResponse bool

	// EncodeRequest enables request body encoding.
	EncodeRequest bool

	// DecodeResponse enables response body decoding.
	DecodeResponse bool

	// EncodeThreshold specifies the request encoding threshold in bytes.
	EncodeThreshold int
}

// NewGlobalRequestConfig creates a new GlobalRequestConfig with default values.
//
// Default timeout is 15 seconds.
// Default EncodeThreshold - 128Kb
// By default client accept encoded response.
func NewGlobalRequestConfig() *GlobalRequestConfig {
	return &GlobalRequestConfig{
		Timeout:               defaultTimeout,
		AcceptEncodedResponse: true,
		DecodeResponse:        true,
		EncodeRequest:         false,
		EncodeThreshold:       defaultEncodeThreshold,
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
	req.acceptEncodedResponse = c.AcceptEncodedResponse
	req.decodeResponse = c.DecodeResponse
	req.encodeRequest = c.EncodeRequest
	req.encodeThreshold = c.EncodeThreshold
}
