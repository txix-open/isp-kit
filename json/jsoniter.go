// Package json provides a high-performance JSON encoder/decoder based on jsoniter.
//
// It extends the standard library with:
//   - Custom time.Time formatting (RFC3339 with milliseconds)
//   - Automatic camelCase conversion for struct field names
//   - Full API compatibility with encoding/json
//
// This package is safe for concurrent use by multiple goroutines.
package json

import (
	"encoding/json"
	"io"

	"github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

// RawMessage is a raw encoded JSON value. It implements Marshaler and
// Unmarshaler and can be used to delay JSON decoding or precompute a JSON
// encoding. It is an alias for encoding/json.RawMessage.
type RawMessage = json.RawMessage

// instance is the shared jsoniter configuration with time and naming extensions.
// nolint:gochecknoglobals
var (
	instance = jsoniter.ConfigDefault
)

// init registers the custom time codec and camelCase naming strategy extensions.
// nolint:gochecknoinits
func init() {
	timeType := reflect2.TypeByName("time.Time")
	tc := NewTimeCodec(FullDateFormat)
	encExt := jsoniter.EncoderExtension{timeType: tc}
	decExt := jsoniter.DecoderExtension{timeType: tc}
	instance.RegisterExtension(encExt)
	instance.RegisterExtension(decExt)

	naming := &namingStrategyExtension{jsoniter.DummyExtension{}, lowerCaseFirstChar}
	instance.RegisterExtension(naming)
}

// Marshal returns the JSON encoding of v.
//
// It uses the default jsoniter configuration with time formatting and camelCase field naming.
func Marshal(v any) ([]byte, error) {
	return instance.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by ptr.
//
// See encoding/json.Unmarshal for details on the conversion.
func Unmarshal(data []byte, ptr any) error {
	return instance.Unmarshal(data, ptr)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return instance.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return instance.NewDecoder(r)
}

// EncodeInto encodes a value into a writer, appending a newline.
//
// This is an optimized function for line-delimited JSON (e.g., NDJSON).
func EncodeInto(w io.Writer, value any) error {
	stream := instance.BorrowStream(w)
	defer instance.ReturnStream(stream)
	stream.WriteVal(value)
	stream.WriteRaw("\n")
	stream.Flush()
	return stream.Error
}
