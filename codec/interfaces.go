package codec

import "io"

type Codec interface {
	Type() string

	EncodeBytes(b []byte) ([]byte, error)
	DecodeBytes(b []byte) ([]byte, error)

	Encode(w io.Writer) (io.WriteCloser, error)
	Decode(rc io.ReadCloser) (io.ReadCloser, error)
}

type decoder interface {
	io.Reader
	Reset(r io.Reader) error
}

type encoder interface {
	io.Writer
	Close() error
	Reset(w io.Writer)
}

type putter interface {
	Put(x any)
}
