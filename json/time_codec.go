package json

import (
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

// FullDateFormat is the default layout for time.Time serialization.
//
// It uses RFC3339 with millisecond precision: "2006-01-02T15:04:05.999Z07:00"
const (
	FullDateFormat = "2006-01-02T15:04:05.999Z07:00"
)

// TimeCodec is a jsoniter extension for custom time.Time encoding and decoding.
//
// It formats time values using a configurable layout string.
type TimeCodec struct {
	format string
}

// NewTimeCodec returns a new TimeCodec with the specified format layout.
//
// The format string follows Go's time.Parse conventions (see time.Layout).
func NewTimeCodec(format string) *TimeCodec {
	return &TimeCodec{format: format}
}

// Decode parses a JSON string into a time.Time value.
//
// It reports an error to the iterator if the string cannot be parsed.
func (codec *TimeCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	t, err := time.Parse(codec.format, iter.ReadString())
	if err != nil {
		iter.ReportError("string -> time.Time", err.Error())
	} else {
		*((*time.Time)(ptr)) = t
	}
}

// IsEmpty checks if a time.Time value is zero.
//
// It returns true if the time is the zero value.
func (codec *TimeCodec) IsEmpty(ptr unsafe.Pointer) bool {
	ts := *((*time.Time)(ptr))
	return ts.IsZero()
}

// Encode writes a time.Time value as a formatted JSON string.
//
// The time is formatted using the codec's layout string.
func (codec *TimeCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(codec.format))
}
