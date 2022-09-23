package httpcli

import (
	"bytes"
	"io"
	"net/http"
)

type Response struct {
	Raw *http.Response

	body []byte
	err  error
	buff *bytes.Buffer
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
	_, err := io.Copy(r.buff, r.Raw.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.body = r.buff.Bytes()
	return r.body, nil
}

// Close
// Release all resources associated with Response (buffer, tcp connection)
// After call, bytes slice returned by Body can not be used
func (r *Response) Close() {
	_ = r.Raw.Body.Close()
	r.body = nil
	releaseBuffer(r.buff)
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode() >= http.StatusOK && r.StatusCode() < http.StatusMultipleChoices
}

func (r *Response) StatusCode() int {
	return r.Raw.StatusCode
}
