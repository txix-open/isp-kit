package json

import (
	"github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

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
