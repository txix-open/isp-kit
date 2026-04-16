package httpcli

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Response wraps an http.Response with buffering and lifecycle management.
type Response struct {
	Raw *http.Response

	cancel context.CancelFunc
	body   []byte
	err    error
	buff   *bytes.Buffer
}

// ReadingResponseMetricHook is a context key for hooks that are called when
// the response body is fully read. Used for metrics collection.
type ReadingResponseMetricHook struct{}

// nolint:gochecknoglobals
var (
	ReadingResponseMetricHookKey = ReadingResponseMetricHook{}
)

// UnsafeBody reads and returns the full response body as bytes.
//
// After calling Close(), the returned data is no longer valid.
// Use BodyCopy() if you need the data after calling Close().
//
// Do not call Close() or modify the returned slice if you want to use
// the data outside the current scope.
func (r *Response) UnsafeBody() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.body != nil {
		return r.body, nil
	}

	defer func() {
		if r.cancel != nil {
			r.cancel() // associated context is no longer needed
			r.cancel = nil
		}
	}()
	_, err := io.Copy(r.buff, r.Raw.Body)
	if err != nil {
		r.err = err
		return nil, err
	}

	if hook, ok := r.Raw.Request.Context().Value(ReadingResponseMetricHookKey).(func()); ok {
		hook()
	}

	r.body = r.buff.Bytes()
	return r.body, nil
}

// Close releases all resources associated with the Response, including buffers,
// TCP connections, and context.
//
// After calling Close(), any bytes slice returned by UnsafeBody() or BodyCopy()
// should not be accessed.
func (r *Response) Close() {
	if r.cancel != nil {
		r.cancel()
		r.cancel = nil
	}
	if r.Raw != nil {
		_ = r.Raw.Body.Close()
	}
	r.body = nil
	releaseBuffer(r.buff)
}

// IsSuccess returns true if the response status code is in the 2xx range.
func (r *Response) IsSuccess() bool {
	return r.StatusCode() >= http.StatusOK && r.StatusCode() < http.StatusMultipleChoices
}

// StatusCode returns the HTTP status code of the response.
func (r *Response) StatusCode() int {
	return r.Raw.StatusCode
}

// BodyCopy returns a copy of the response body as bytes.
//
// The returned slice remains valid even after calling Close().
func (r *Response) BodyCopy() ([]byte, error) {
	body, err := r.UnsafeBody()
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(body))
	copy(copied, body)

	return copied, nil
}
