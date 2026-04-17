package httpcli

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/txix-open/isp-kit/json"
)

// RequestBodyWriter defines the interface for writing request bodies.
type RequestBodyWriter interface {
	Write(req *http.Request, w io.Writer) error
}

// jsonRequest handles JSON-encoded request bodies.
type jsonRequest struct {
	value any
}

// Write sets the Content-Type header to application/json and encodes the value as JSON.
func (j jsonRequest) Write(req *http.Request, w io.Writer) error {
	req.Header.Set("Content-Type", "application/json")
	return json.EncodeInto(w, j.value)
}

// plainRequest handles raw byte request bodies.
type plainRequest struct {
	value []byte
}

// Write writes the raw bytes to the request body.
func (p plainRequest) Write(_ *http.Request, w io.Writer) error {
	_, err := w.Write(p.value)
	return err
}

// formRequest handles form-encoded request bodies.
type formRequest struct {
	data map[string][]string
}

// Write sets the Content-Type header to application/x-www-form-urlencoded
// and encodes the data as form parameters.
func (f formRequest) Write(req *http.Request, w io.Writer) error {
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	data := url.Values(f.data).Encode()
	_, err := w.Write([]byte(data))
	return err
}

// ResponseBodyReader defines the interface for reading response bodies.
type ResponseBodyReader interface {
	Read(r io.Reader) error
}

// jsonResponse handles JSON-encoded response bodies.
type jsonResponse struct {
	ptr any
}

// Read unmarshals the response body into the target pointer.
// Uses optimized unmarshaling when the source is a bytes.Buffer.
func (j jsonResponse) Read(r io.Reader) error {
	buff, isBuff := r.(*bytes.Buffer)
	if isBuff {
		return json.Unmarshal(buff.Bytes(), j.ptr)
	}
	return json.NewDecoder(r).Decode(j.ptr)
}
