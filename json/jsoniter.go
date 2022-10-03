package json

import (
	"encoding/json"
	"io"

	"github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

type RawMessage json.RawMessage

var (
	instance = jsoniter.ConfigDefault
)

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

func Marshal(v interface{}) ([]byte, error) {
	return instance.Marshal(v)
}

func Unmarshal(data []byte, ptr interface{}) error {
	return instance.Unmarshal(data, ptr)
}

func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return instance.NewEncoder(w)
}

func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return instance.NewDecoder(r)
}

func EncodeInto(w io.Writer, value any) error {
	stream := instance.BorrowStream(w)
	defer instance.ReturnStream(stream)
	stream.WriteVal(value)
	stream.WriteRaw("\n")
	stream.Flush()
	return stream.Error
}
