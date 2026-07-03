package endpoint

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/codec"
	http2 "github.com/txix-open/isp-kit/http"
)

type DecodeCodec interface {
	Type() string
	Decode(r io.ReadCloser) (io.ReadCloser, error)
}

const (
	defaultEncodeThreshold = codec.DefaultEncodeThreshold
)

// DecodeRequest returns middleware that transparently handles encoded requests.
//
// If Content-Encoding matches codec, request body is decoded.
//
// Important:
//   - Should be placed before logging middleware if logs must see decoded data.
func DecodeRequest(codec DecodeCodec) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if r.Header.Get("Content-Encoding") == codec.Type() {
				decoded, err := codec.Decode(r.Body)
				if err != nil {
					return err
				}

				r.Body = decoded
				r.Header.Del("Content-Encoding")
				r.Header.Del("Content-Length")
			}
			return next(ctx, w, r)
		}
	}
}

type EncodeCodec interface {
	Type() string
	Encode(w io.Writer) (io.WriteCloser, error)
}

// EncodeResponse returns middleware that transparently handles response encoding.
//
// If client supports Accept-Encoding and response body size > encodeThresholdBytes,
// the response is encoded using codec.
//
// The middleware finalizes the response after the handler returns. It may
// implicitly call WriteHeader(http.StatusOK) and flush buffered response data
// even if the handler never calls Write or WriteHeader.
//
// Important:
//   - Should be placed before logging middleware if logs must see plain data.
func EncodeResponse(codec EncodeCodec, encodeThresholdBytes int) http2.Middleware {
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if !acceptsEncoding(r.Header.Get("Accept-Encoding"), codec.Type()) {
				return next(ctx, w, r)
			}

			cw := newEncoderWriter(w, codec, encodeThresholdBytes)

			err := next(ctx, cw, r)
			closeErr := cw.Close()

			if err != nil {
				return err
			}
			return closeErr
		}
	}
}

type encoderWriter struct {
	http.ResponseWriter

	codec EncodeCodec

	buffer          bytes.Buffer
	encodeThreshold int

	writer io.WriteCloser
	status int
}

func newEncoderWriter(w http.ResponseWriter, codec EncodeCodec, encodeThreshold int) *encoderWriter {
	return &encoderWriter{
		ResponseWriter:  w,
		codec:           codec,
		encodeThreshold: encodeThreshold,
	}
}

func (w *encoderWriter) WriteHeader(code int) {
	if w.status != 0 {
		return
	}
	w.status = code
}

func (w *encoderWriter) Write(b []byte) (int, error) {
	if w.writer != nil {
		return w.writer.Write(b)
	}

	// fast path: big first write
	if w.buffer.Len() == 0 && len(b) > w.encodeThreshold {
		return w.startEncoding(b)
	}

	n, err := w.buffer.Write(b)
	if err != nil {
		return n, err
	}

	if w.buffer.Len() > w.encodeThreshold {
		data := append([]byte(nil), w.buffer.Bytes()...)
		w.buffer.Reset()

		return w.startEncoding(data)
	}

	return n, nil
}

func (w *encoderWriter) Close() error {
	if w.writer != nil {
		return w.writer.Close()
	}

	status := w.status
	if status == 0 {
		status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(status)

	_, err := w.ResponseWriter.Write(w.buffer.Bytes())
	return err
}

// Hijack implements http.Hijacker.
//
// Hijacking is only allowed before the response has started (no buffered or
// encoded data). Once any response data has been written or encoding has been
// triggered, hijacking is disallowed to prevent losing buffered state and to
// preserve response integrity.
func (w *encoderWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.writer != nil || w.buffer.Len() > 0 || w.status != 0 {
		return nil, nil, errors.New("encoderWriter: hijack not allowed after response started")
	}

	upstream, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("encoderWriter: upstream writer doesn't implement Hijack")
	}

	return upstream.Hijack()
}

func (w *encoderWriter) startEncoding(b []byte) (int, error) {
	writer, err := w.codec.Encode(w.ResponseWriter)
	if err != nil {
		return 0, err
	}
	w.writer = writer

	h := w.Header()
	h.Set("Content-Encoding", w.codec.Type())
	h.Add("Vary", "Accept-Encoding")
	h.Del("Content-Length")

	status := w.status
	if status == 0 {
		status = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(status)

	return w.writer.Write(b)
}

func acceptsEncoding(header, encoding string) bool {
	for part := range strings.SplitSeq(header, ",") {
		if strings.TrimSpace(part) == encoding {
			return true
		}
	}

	return false
}
