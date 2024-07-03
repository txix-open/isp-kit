package httpcli

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type Response struct {
	Raw *http.Response

	cancel context.CancelFunc
	body   []byte
	err    error
	buff   *bytes.Buffer
}

// Body
// Read and return full response body
// Be careful, after calling Close returned data is no longer available
// Do not call close or copy slice if you want to use data outside the calling function
func (r *Response) Body() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.body != nil {
		return r.body, nil
	}
	defer func() {
		if r.cancel != nil {
			r.cancel() //associated context is no longer needed
			r.cancel = nil
		}
	}()
	_, err := io.Copy(r.buff, r.Raw.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.body = r.buff.Bytes()
	return r.body, nil
}

// Close
// Release all resources associated with Response (buffer, tcp connection, context)
// After call, bytes slice returned by Body can not be used
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

func (r *Response) IsSuccess() bool {
	return r.StatusCode() >= http.StatusOK && r.StatusCode() < http.StatusMultipleChoices
}

func (r *Response) StatusCode() int {
	return r.Raw.StatusCode
}

// BodyCopy
// Return copy of response body
// Slice is available after calling Close
func (r *Response) BodyCopy() ([]byte, error) {
	body, err := r.Body()
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(body))
	copy(copied, body)

	return copied, nil
}
