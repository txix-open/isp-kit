package buffer

import (
	"bytes"
)

type RequestBody struct {
	body *bytes.Buffer
}

func NewRequestBody(body []byte) RequestBody {
	return RequestBody{
		body: bytes.NewBuffer(body),
	}
}

func (r RequestBody) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r RequestBody) Close() error {
	return nil
}
