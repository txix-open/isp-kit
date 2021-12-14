package json

import (
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

const (
	FullDateFormat = "2006-01-02T15:04:05.999Z07:00"
)

type TimeCodec struct {
	format string
}

func NewTimeCodec(format string) *TimeCodec {
	return &TimeCodec{format: format}
}

func (codec *TimeCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	t, err := time.Parse(codec.format, iter.ReadString())
	if err != nil {
		iter.ReportError("string -> time.Time", err.Error())
	} else {
		*((*time.Time)(ptr)) = t
	}
}

func (codec *TimeCodec) IsEmpty(ptr unsafe.Pointer) bool {
	ts := *((*time.Time)(ptr))
	return ts.IsZero()
}

func (codec *TimeCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(codec.format))
}
