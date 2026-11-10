// Package file provides log file output with rotation using lumberjack.
package file

import (
	"io"
	"net/url"
	"strconv"

	"github.com/txix-open/isp-kit/errors"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	lumberjackSchema = "lumberjack"
)

// nolint:gochecknoinits
func init() {
	err := zap.RegisterSink(lumberjackSchema, func(u *url.URL) (zap.Sink, error) {
		r, err := fileOutputConfigFromUrl(u)
		if err != nil {
			return nil, errors.WithMessage(err, "unmarshal rotation config")
		}
		return newSink(NewFileWriter(*r)), nil
	})
	if err != nil {
		panic(err)
	}
}

// Output configures log file rotation.
type Output struct {
	// File is the path to the log file.
	File string
	// MaxSizeMb is the maximum file size in megabytes before rotation.
	// Used when PartSizeKb is not set.
	MaxSizeMb int
	// PartSizeKb is the size of a log part in kilobytes. When set (> 0)
	// it takes precedence over MaxSizeMb and enables byte-precise
	// rotation limits (e.g. 218 for VictoriaLogs).
	PartSizeKb int
	// MaxDays is the maximum number of days to retain old log files.
	// Used only when PartSizeKb is not set.
	MaxDays int
	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int
	// Compress enables compression of rotated log files.
	Compress bool
}

// ConfigToUrl converts an Output configuration to a URL.
func ConfigToUrl(r Output) *url.URL {
	values := url.Values{
		"file":       {r.File},
		"maxSizeMb":  {strconv.Itoa(r.MaxSizeMb)},
		"partSizeKb": {strconv.Itoa(r.PartSizeKb)},
		"maxDays":    {strconv.Itoa(r.MaxDays)},
		"maxBackups": {strconv.Itoa(r.MaxBackups)},
		"compress":   {strconv.FormatBool(r.Compress)},
	}
	u := &url.URL{
		Scheme:   lumberjackSchema,
		RawQuery: values.Encode(),
	}
	return u
}

// fileOutputConfigFromUrl parses an Output configuration from a URL.
func fileOutputConfigFromUrl(u *url.URL) (*Output, error) {
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil, errors.WithMessage(err, "parse lumberjack params")
	}
	file := values.Get("file")
	maxSizeMb, err := strconv.Atoi(values.Get("maxSizeMb"))
	if err != nil {
		return nil, errors.WithMessage(err, "parse maxSizeMb")
	}
	// partSizeKb is parsed leniently to stay compatible with URLs built
	// before the parameter was introduced.
	partSizeKb := 0
	if v := values.Get("partSizeKb"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			partSizeKb = n
		}
	}
	maxDays, err := strconv.Atoi(values.Get("maxDays"))
	if err != nil {
		return nil, errors.WithMessage(err, "parse maxDays")
	}
	maxBackups, err := strconv.Atoi(values.Get("maxBackups"))
	if err != nil {
		return nil, errors.WithMessage(err, "parse maxBackups")
	}
	compress, err := strconv.ParseBool(values.Get("compress"))
	if err != nil {
		return nil, errors.WithMessage(err, "parse compress")
	}
	return &Output{
		File:       file,
		MaxSizeMb:  maxSizeMb,
		PartSizeKb: partSizeKb,
		MaxDays:    maxDays,
		MaxBackups: maxBackups,
		Compress:   compress,
	}, nil
}

// NewFileWriter creates a new log file writer with rotation.
// When PartSizeKb is set a byte-precise rotation writer is used, because
// lumberjack only supports whole-megabyte size limits.
func NewFileWriter(r Output) io.WriteCloser {
	if r.PartSizeKb > 0 {
		return newRotatingFileWriter(r.File, int64(r.PartSizeKb)*kilobytes, r.MaxBackups, r.Compress)
	}
	return &lumberjack.Logger{
		Filename:   r.File,
		MaxSize:    r.MaxSizeMb,
		MaxAge:     r.MaxDays,
		MaxBackups: r.MaxBackups,
		Compress:   r.Compress,
	}
}

type sink struct {
	io.WriteCloser
}

// Sync flushes buffered writes.
func (s sink) Sync() error {
	return s.Close()
}

// newSink wraps a writer as a Zap sink.
func newSink(writer io.WriteCloser) sink {
	return sink{
		WriteCloser: writer,
	}
}
