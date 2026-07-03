package codec

import (
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// Level is an alias for zstd.EncoderLevel representing encoder compression levels.
type ZstdLevel zstd.EncoderLevel

const (
	ZstdSpeedFastest           = zstd.SpeedFastest
	ZstdSpeedDefault           = zstd.SpeedDefault
	ZstdSpeedBetterCompression = zstd.SpeedBetterCompression
	ZstdSpeedBestCompression   = zstd.SpeedBestCompression
)

type ZstdConfig struct {
	Level ZstdLevel
}

type Zstd struct {
	encoderPool sync.Pool
	decoderPool sync.Pool
}

// NewZstd create instance of zstd
func NewZstd(cfg ZstdConfig) *Zstd {
	encoderLevel := zstd.EncoderLevel(cfg.Level)
	return &Zstd{
		encoderPool: sync.Pool{
			New: func() any {
				enc, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(encoderLevel))
				return enc
			},
		},
		decoderPool: sync.Pool{
			New: func() any {
				dec, _ := zstd.NewReader(nil)
				return dec
			},
		},
	}
}

func (*Zstd) Type() string {
	return "zstd"
}

// Encode data using zstd.
func (m *Zstd) Encode(w io.Writer) (io.WriteCloser, error) {
	encoder := m.encoderPool.Get().(*zstd.Encoder) //nolint:forcetypeassert

	// Reset encoder to new destination
	encoder.Reset(w)

	return &writeCloser{
		encoder: encoder,
		pool:    &m.encoderPool,
	}, nil
}

// EncodeBytes encode data using zstd.
func (m *Zstd) EncodeBytes(data []byte) ([]byte, error) {
	encoder := m.encoderPool.Get().(*zstd.Encoder) //nolint:forcetypeassert
	defer m.encoderPool.Put(encoder)

	return encoder.EncodeAll(data, nil), nil
}

// Decode wraps body with a zstd decoder.
func (m *Zstd) Decode(body io.ReadCloser) (io.ReadCloser, error) {
	decoder := m.decoderPool.Get().(*zstd.Decoder) //nolint:forcetypeassert

	err := decoder.Reset(body)
	if err != nil {
		m.decoderPool.Put(decoder)
		return nil, err
	}

	return &readCloser{
		decoder: decoder,
		pool:    &m.decoderPool,
		body:    body,
	}, nil
}

// DecodeBytes decodes bytes body using zstd.
func (m *Zstd) DecodeBytes(data []byte) ([]byte, error) {
	decoder := m.decoderPool.Get().(*zstd.Decoder) //nolint:forcetypeassert
	defer m.decoderPool.Put(decoder)

	return decoder.DecodeAll(data, nil)
}
