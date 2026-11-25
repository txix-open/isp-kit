package log

import (
	"fmt"
	"strconv"
	"unicode/utf8"
	"uuid"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Field names added to the root of every part of a split JSON log record. A
// receiver groups parts by [SplitFieldPackageId], sorts them by
// [SplitFieldPackageIndex] and checks that their number is
// [SplitFieldPackageTotal].
const (
	// SplitFieldPackageId is the id shared by all parts of one original record.
	SplitFieldPackageId = "logPackageId"
	// SplitFieldPackageIndex is the 1-based number of a part in a package.
	SplitFieldPackageIndex = "logPackageIndex"
	// SplitFieldPackageTotal is the total number of parts in a package.
	SplitFieldPackageTotal = "logPackageTotal"
)

const (
	// DefaultMaxLineKb is the default limit of a part in kilobytes. It is
	// 218 KiB and leaves a margin up to the 256 KB limit of a log collector.
	DefaultMaxLineKb = 218
	// DefaultMaxLineBytes is [DefaultMaxLineKb] in bytes.
	DefaultMaxLineBytes = DefaultMaxLineKb * 1024
	// probePartNumber is the maximum part number reserved when the overhead of
	// a part is measured. It has 9 digits.
	probePartNumber = 999999999
	// metaFieldsCapacity is the maximum number of package meta fields of a part.
	metaFieldsCapacity = 4
	// escapedSpecialWidth is the size of a character escaped with a backslash.
	escapedSpecialWidth = 2
	// escapedControlWidth is the size of a control character escaped as \u00XX.
	escapedControlWidth = 6
	// controlByteBoundary is the first printable ASCII byte.
	controlByteBoundary = 0x20
	// jsonBraces is the number of braces around an encoded field.
	jsonBraces = 2
	// fieldKeySyntaxWidth is the size of the comma, quotes and colon around a
	// field key: ,"key": without the key itself.
	fieldKeySyntaxWidth = 4
	// fieldValueQuotesWidth is the size of the quotes around a string value.
	fieldValueQuotesWidth = 2
	// bytesPerKb is the number of bytes in a kilobyte.
	bytesPerKb = 1024
	// minSplitParts is the minimum number of parts of a split record: a record
	// that fits into one part is written as is.
	minSplitParts = 2
	// utf8ContinuationMask and utf8ContinuationPrefix identify a continuation
	// byte of a UTF-8 sequence: 10xxxxxx.
	utf8ContinuationMask   = 0xC0
	utf8ContinuationPrefix = 0x80
	// decimalBase is the base of the decimal representation of an integer.
	decimalBase = 10
)

// errCannotSplit reports that a record cannot be split without losing data.
// Such a record is written as is, see [splittingCore.Write].
var errCannotSplit = errors.New("cannot split log record")

// SplitConfig configures splitting of oversized JSON log records into parts.
//
// When enabled, a record whose JSON representation is longer than
// [SplitConfig.MaxLineKb] kilobytes is written as several ordinary records.
// Every part keeps the original level, time and field names, and additionally
// carries [SplitFieldPackageId], [SplitFieldPackageIndex] and
// [SplitFieldPackageTotal]. The request id, if it was in the record, and the
// message are written to every part as well.
//
// Fields that fit into a part are packed into it unchanged. A field that does
// not fit keeps its name and its value is cut into pieces: concatenation of the
// pieces in the part order restores the value. Only string fields are cut.
//
// Splitting is ignored in development mode, where the output is not JSON.
type SplitConfig struct {
	// Enabled turns splitting on. Splitting is off by default.
	Enabled bool
	// MaxLineKb is the maximum size of one part in kilobytes. A value less than
	// or equal to zero makes splitting inactive.
	MaxLineKb int
}

// enabled reports whether the configuration requires splitting.
func (c SplitConfig) enabled() bool {
	return c.Enabled && c.MaxLineKb > 0
}

// maxLineBytes returns the byte limit of one part.
func (c SplitConfig) maxLineBytes() int {
	return c.MaxLineKb * bytesPerKb
}

// splittingCore writes JSON records to a sink and splits oversized records into
// several records.
//
// The core encodes a record once. A record that fits is written to the sink as
// is, without a second encoding. An oversized record is planned and every part
// is encoded and written separately. Every part carries the same level, time,
// message and logger fields, so the parts can be filtered by them.
//
// The core is assembled manually, see [newSplitLogger]: a regular zap core
// encodes and writes a record itself, so a splitting wrapper would have to
// encode every record twice.
//
// The core is safe for concurrent use: zap encoders encode an entry into a new
// buffer and keep no state between calls.
type splittingCore struct {
	encoder zapcore.Encoder
	sink    zapcore.WriteSyncer
	level   zapcore.LevelEnabler
	// bareEncConfig is the encoder config without the record envelope: it is
	// used to measure a single field.
	bareEncConfig zapcore.EncoderConfig
	maxLineBytes  int
	// withRequestId reports whether the request id is attached to the logger by
	// With. The encoders write it into every record, so a meta request id field
	// is not added.
	withRequestId bool
	// withNamespace reports whether a namespace is attached to the logger by
	// With. The meta fields of a part would be nested into that namespace, so
	// such records are written as is.
	withNamespace bool
}

// newSplittingCore creates a core that writes ready lines to sink.
func newSplittingCore(
	encoder zapcore.Encoder,
	sink zapcore.WriteSyncer,
	level zapcore.LevelEnabler,
	encConfig zapcore.EncoderConfig,
	cfg SplitConfig,
) *splittingCore {
	return &splittingCore{
		encoder:       encoder,
		sink:          sink,
		level:         level,
		bareEncConfig: fieldOnlyEncoderConfig(encConfig),
		maxLineBytes:  cfg.maxLineBytes(),
	}
}

// fieldOnlyEncoderConfig returns a copy of cfg without the record envelope: a
// field is measured on its own, so no level, time or message is written.
func fieldOnlyEncoderConfig(cfg zapcore.EncoderConfig) zapcore.EncoderConfig {
	cfg.LevelKey = ""
	cfg.TimeKey = ""
	cfg.MessageKey = ""
	cfg.NameKey = ""
	cfg.CallerKey = ""
	cfg.FunctionKey = ""
	cfg.StacktraceKey = ""
	return cfg
}

// Enabled reports whether the given level is enabled.
func (c *splittingCore) Enabled(level zapcore.Level) bool {
	return c.level.Enabled(level)
}

// With returns a new core with the given fields attached to the encoder. The
// request id and namespaces are tracked: they change how the parts are built.
func (c *splittingCore) With(fields []zapcore.Field) zapcore.Core {
	if len(fields) == 0 {
		return c
	}

	encoder := c.encoder.Clone()
	for i := range fields {
		fields[i].AddTo(encoder)
	}

	clone := *c
	clone.encoder = encoder
	clone.withRequestId = c.withRequestId || hasField(fields, requestid.LogKey)
	clone.withNamespace = c.withNamespace || hasNamespaceField(fields)
	return &clone
}

// Check registers the core itself.
func (c *splittingCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return checkedEntry
	}
	return checkedEntry.AddCore(entry, c)
}

// Sync flushes the sink.
func (c *splittingCore) Sync() error {
	return c.sink.Sync()
}

// Write encodes a record once and writes it to the sink. An oversized record is
// split into parts: every part is encoded and written separately. A record that
// cannot be split is written as is, so a logging call never loses data because
// of splitting.
func (c *splittingCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// Escaping only makes a record longer, so the raw content is a lower bound
	// of the encoded record. A record that is already longer than the limit does
	// not need to be encoded before splitting.
	if c.definitelyOversized(entry, fields) {
		return c.writeSplit(entry, fields, nil)
	}

	line, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}
	if recordLen(line) <= c.maxLineBytes {
		return c.writeBuffer(line)
	}
	return c.writeSplit(entry, fields, line)
}

// writeSplit plans the parts of an oversized record and writes them. The
// encoded whole record is passed when it is already available: it is written as
// is when the record cannot be split.
func (c *splittingCore) writeSplit(
	entry zapcore.Entry,
	fields []zapcore.Field,
	oversizedLine *buffer.Buffer,
) error {
	if c.withNamespace {
		return c.writeWhole(entry, fields, oversizedLine)
	}

	sizes := make([]int, len(fields))
	if !c.measureFields(fields, sizes) {
		return c.writeWhole(entry, fields, oversizedLine)
	}

	logPackageId := uuid.New().String()

	packable, packableSizes, requestIdField, hasRequestId := c.extractRequestId(fields, sizes)
	parts := c.planParts(entry, packable, packableSizes, logPackageId, requestIdField, hasRequestId)
	if len(parts) < minSplitParts {
		return c.writeWhole(entry, fields, oversizedLine)
	}

	if oversizedLine != nil {
		oversizedLine.Free()
	}
	return c.writeParts(entry, parts, logPackageId, requestIdField, hasRequestId)
}

// writeParts encodes and writes every part with the same encoder as a whole
// record. It stops at the first error: the remaining parts are not written
// when the encoder or the sink has failed.
func (c *splittingCore) writeParts(
	entry zapcore.Entry,
	parts [][]zapcore.Field,
	logPackageId string,
	requestIdField zapcore.Field,
	hasRequestId bool,
) error {
	for i, part := range parts {
		partFields := make([]zapcore.Field, 0, len(part)+metaFieldsCapacity)
		partFields = append(partFields,
			metaFields(logPackageId, i+1, len(parts), requestIdField, hasRequestId)...)
		partFields = append(partFields, part...)

		line, err := c.encoder.EncodeEntry(entry, partFields)
		if err != nil {
			return errors.WithMessage(err, "cannot encode log part")
		}
		err = c.writeBuffer(line)
		if err != nil {
			return errors.WithMessage(err, "cannot write log part")
		}
	}
	return nil
}

// writeWhole writes the record as is. The encoded line is reused when it is
// already available.
func (c *splittingCore) writeWhole(
	entry zapcore.Entry,
	fields []zapcore.Field,
	oversizedLine *buffer.Buffer,
) error {
	if oversizedLine != nil {
		return c.writeBuffer(oversizedLine)
	}

	line, err := c.encoder.EncodeEntry(entry, fields)
	if err != nil {
		return err
	}
	return c.writeBuffer(line)
}

// writeBuffer writes the encoded line to the sink and releases the buffer.
func (c *splittingCore) writeBuffer(line *buffer.Buffer) error {
	defer line.Free()

	_, err := c.sink.Write(line.Bytes())
	return err
}

// definitelyOversized reports whether the raw content of the record is already
// longer than the limit. Every byte takes at least one byte in the JSON output,
// so such a record is definitely oversized.
func (c *splittingCore) definitelyOversized(entry zapcore.Entry, fields []zapcore.Field) bool {
	total := len(entry.Message)
	if total > c.maxLineBytes {
		return true
	}

	for i := range fields {
		field := fields[i]
		if field.Type == zapcore.StringType {
			total += len(field.Key) + len(field.String)
		} else if field.Type == zapcore.ByteStringType {
			value, ok := field.Interface.([]byte)
			if ok {
				total += len(field.Key) + len(value)
			}
		}
		if total > c.maxLineBytes {
			return true
		}
	}
	return false
}

// measureFields fills sizes with the exact size of every field in a record. It
// reports false when a field cannot be measured field by field.
func (c *splittingCore) measureFields(fields []zapcore.Field, sizes []int) bool {
	sizer := &fieldSizer{encConfig: c.bareEncConfig, limit: c.maxLineBytes}
	for i := range fields {
		size, err := sizer.size(fields[i])
		if err != nil {
			return false
		}
		sizes[i] = size
	}
	return true
}

// extractRequestId removes the request id from the packable fields: it is
// written to every part by [metaFields]. When the request id was attached to
// the logger by With, the encoders already write it into every part, so no meta
// field is added and the duplicate is removed. sizes stays aligned with the
// returned fields.
func (c *splittingCore) extractRequestId(
	fields []zapcore.Field,
	sizes []int,
) ([]zapcore.Field, []int, zapcore.Field, bool) {
	requestIdField, hasRequestId := findRequestId(fields)
	if !hasRequestId {
		return fields, sizes, zapcore.Field{}, false
	}
	packable, packableSizes := withoutField(fields, sizes, requestid.LogKey)
	if c.withRequestId {
		return packable, packableSizes, zapcore.Field{}, false
	}
	return packable, packableSizes, requestIdField, true
}

// planParts packs the fields into parts. Small fields are packed greedily in the
// original order, a field that does not fit into a part is cut into pieces and
// goes to its own parts. sizes contains the exact size of every field. Every
// part carries the whole envelope (level, time, message and logger fields), so
// all parts share the same budget. It returns nil when the record cannot be
// split: the limit is smaller than the envelope (for example, an oversized
// message), a field cannot be packed or a field cannot be cut.
func (c *splittingCore) planParts(
	entry zapcore.Entry,
	fields []zapcore.Field,
	sizes []int,
	logPackageId string,
	requestIdField zapcore.Field,
	hasRequestId bool,
) [][]zapcore.Field {
	budget, ok := c.partBudget(entry, logPackageId, requestIdField, hasRequestId)
	if !ok {
		return nil
	}

	parts := make([][]zapcore.Field, 0, len(fields)+1)
	current := make([]zapcore.Field, 0, len(fields))
	currentSize := 0

	for i, field := range fields {
		if field.Type == zapcore.SkipType {
			continue // the field writes nothing, it must not create a part
		}
		size := sizes[i]
		if size > budget {
			if len(current) > 0 {
				parts = append(parts, current)
				current, currentSize = nil, 0
			}
			chunks := c.chunkField(field, size, budget)
			if chunks == nil {
				return nil
			}
			parts = append(parts, chunks...)
			continue
		}
		if currentSize+size > budget {
			if len(current) > 0 {
				parts = append(parts, current)
			}
			current, currentSize = nil, 0
		}
		current = append(current, field)
		currentSize += size
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}

// partBudget returns the free space of a part. It reports false when the limit
// is smaller than the envelope, for example when the message itself does not fit
// into a part.
func (c *splittingCore) partBudget(
	entry zapcore.Entry,
	logPackageId string,
	requestIdField zapcore.Field,
	hasRequestId bool,
) (int, bool) {
	baseline, err := c.partBaseline(entry, logPackageId, requestIdField, hasRequestId)
	if err != nil {
		return 0, false
	}
	budget := c.maxLineBytes - baseline
	if budget <= 0 {
		return 0, false
	}
	return budget, true
}

// partBaseline returns the size of a part without user fields: the envelope,
// the fields of the logger and the package meta fields. The meta fields are
// measured with the maximum 9-digit part numbers, so the baseline is an upper
// bound for any real part.
func (c *splittingCore) partBaseline(
	entry zapcore.Entry,
	logPackageId string,
	requestIdField zapcore.Field,
	hasRequestId bool,
) (int, error) {
	probeFields := metaFields(logPackageId, probePartNumber, probePartNumber, requestIdField, hasRequestId)
	line, err := c.encoder.EncodeEntry(entry, probeFields)
	if err != nil {
		return 0, err
	}
	defer line.Free()

	return recordLen(line), nil
}

// chunkField cuts a string field into pieces that fit into a part. The field
// keeps its name. It returns nil when the field cannot be cut without losing
// data: it is not a string field, it contains invalid UTF-8 or it is encoded
// with additional fields (for example, an error with causes).
func (c *splittingCore) chunkField(field zapcore.Field, fieldSize int, budget int) [][]zapcore.Field {
	if !utf8.ValidString(field.Key) {
		return nil
	}
	value, ok := stringValue(field)
	if !ok || !utf8.ValidString(value) {
		return nil
	}

	keyOverhead := fieldKeyOverhead(field.Key)
	// A string field is encoded as a JSON string by definition. An error and a
	// stringer field are encoded as a string only when their size matches the
	// key and the escaped value: an error with causes, a verbose message or a
	// panicking stringer adds extra fields.
	if (field.Type == zapcore.ErrorType || field.Type == zapcore.StringerType) &&
		fieldSize != keyOverhead+fieldValueQuotesWidth+escapedWidth(value) {
		return nil
	}

	valueBudget := budget - keyOverhead - fieldValueQuotesWidth
	if valueBudget <= 0 {
		return nil
	}
	chunks, ok := chunkByBytes(value, valueBudget)
	if !ok {
		return nil
	}

	parts := make([][]zapcore.Field, 0, len(chunks))
	for _, chunk := range chunks {
		parts = append(parts, []zapcore.Field{zap.String(field.Key, chunk)})
	}
	return parts
}

// fieldSizer returns the size of a field without encoding it when the type
// allows, and falls back to a measuring encoder otherwise.
type fieldSizer struct {
	encConfig zapcore.EncoderConfig
	encoder   zapcore.Encoder
	// limit is the limit of the whole record. A string field longer than limit
	// cannot fit into a record and is not measured.
	limit int
}

// size returns the size of the field in a record, including the separating
// comma.
func (s *fieldSizer) size(field zapcore.Field) (int, error) {
	if size, ok := oversizedStringSize(field, s.limit); ok {
		return size, nil
	}
	if size, ok := cheapFieldSize(field); ok {
		return size, nil
	}
	if s.encoder == nil {
		s.encoder = zapcore.NewJSONEncoder(s.encConfig)
	}
	return fieldSize(s.encoder, field)
}

// fieldSize returns the size of a field in a record, including the separating
// comma, by encoding it on its own. A namespace field has no size of its own:
// it changes the structure of the record and cannot be packed field by field.
func fieldSize(encoder zapcore.Encoder, field zapcore.Field) (int, error) {
	if field.Type == zapcore.SkipType {
		return 0, nil
	}
	if field.Type == zapcore.NamespaceType {
		return 0, errCannotSplit
	}

	line, err := encoder.EncodeEntry(zapcore.Entry{}, []zapcore.Field{field})
	if err != nil {
		return 0, err
	}
	defer line.Free()

	// The encoded field is {"key":value}; the braces are not a part of the
	// field, the comma before it is.
	size := recordLen(line) - jsonBraces
	if size < 0 {
		return 0, errCannotSplit
	}
	return size + 1, nil
}

// oversizedStringSize reports whether a string field alone is longer than the
// whole record limit. The returned size is a placeholder larger than any part
// budget: only the fact that the field does not fit is used.
func oversizedStringSize(field zapcore.Field, limit int) (int, bool) {
	if limit <= 0 {
		return 0, false
	}
	if field.Type == zapcore.StringType {
		if len(field.String) > limit {
			return limit + 1, true
		}
		return 0, false
	}
	if field.Type == zapcore.ByteStringType {
		value, ok := field.Interface.([]byte)
		if ok && len(value) > limit {
			return limit + 1, true
		}
	}
	return 0, false
}

// cheapFieldSize returns the exact size of a field without encoding it. It
// supports the types with a known JSON representation: skip, strings, byte
// strings, booleans and integers. Invalid UTF-8 and other types are measured by
// an encoder.
func cheapFieldSize(field zapcore.Field) (int, bool) {
	if !utf8.ValidString(field.Key) {
		return 0, false
	}
	keyOverhead := fieldKeyOverhead(field.Key)
	if field.Type == zapcore.SkipType {
		return 0, true
	}
	if field.Type == zapcore.StringType {
		if !utf8.ValidString(field.String) {
			return 0, false
		}
		return keyOverhead + fieldValueQuotesWidth + escapedWidth(field.String), true
	}
	if field.Type == zapcore.ByteStringType {
		value, ok := field.Interface.([]byte)
		if !ok || !utf8.Valid(value) {
			return 0, false
		}
		return keyOverhead + fieldValueQuotesWidth + escapedBytesWidth(value), true
	}
	if field.Type == zapcore.BoolType {
		if field.Integer == 1 {
			return keyOverhead + len("true"), true
		}
		return keyOverhead + len("false"), true
	}
	if isIntFieldType(field.Type) {
		return keyOverhead + intWidth(field.Integer), true
	}
	if isUintFieldType(field.Type) {
		// The unsigned value is stored in the signed integer field as a bit
		// pattern, see [zapcore.Field].
		return keyOverhead + uintWidth(uint64(field.Integer)), true //nolint:gosec
	}
	return 0, false
}

// isIntFieldType reports whether the field type encodes a signed integer.
func isIntFieldType(fieldType zapcore.FieldType) bool {
	return fieldType == zapcore.Int64Type ||
		fieldType == zapcore.Int32Type ||
		fieldType == zapcore.Int16Type ||
		fieldType == zapcore.Int8Type
}

// isUintFieldType reports whether the field type encodes an unsigned integer.
func isUintFieldType(fieldType zapcore.FieldType) bool {
	return fieldType == zapcore.Uint64Type ||
		fieldType == zapcore.Uint32Type ||
		fieldType == zapcore.Uint16Type ||
		fieldType == zapcore.Uint8Type ||
		fieldType == zapcore.UintptrType
}

// intWidth returns the number of decimal digits of value with the sign.
func intWidth(value int64) int {
	var buff [20]byte
	return len(strconv.AppendInt(buff[:0], value, decimalBase))
}

// uintWidth returns the number of decimal digits of value.
func uintWidth(value uint64) int {
	var buff [20]byte
	return len(strconv.AppendUint(buff[:0], value, decimalBase))
}

// metaFields returns the fields that identify a part in a package. The request
// id goes first and the package fields follow it, so the root of a part carries
// level, time and request id, and then the package metadata.
func metaFields(
	logPackageId string,
	index int,
	total int,
	requestIdField zapcore.Field,
	hasRequestId bool,
) []zapcore.Field {
	fields := make([]zapcore.Field, 0, metaFieldsCapacity)
	if hasRequestId {
		fields = append(fields, requestIdField)
	}
	fields = append(fields,
		zap.String(SplitFieldPackageId, logPackageId),
		zap.Int(SplitFieldPackageIndex, index),
		zap.Int(SplitFieldPackageTotal, total),
	)
	return fields
}

// findRequestId returns the request id from the log fields.
func findRequestId(fields []zapcore.Field) (zapcore.Field, bool) {
	for i := range fields {
		if fields[i].Key == requestid.LogKey && fields[i].Type == zapcore.StringType {
			return fields[i], true
		}
	}
	return zapcore.Field{}, false
}

// hasField reports whether the fields contain a field with the given key.
func hasField(fields []zapcore.Field, key string) bool {
	for i := range fields {
		if fields[i].Key == key {
			return true
		}
	}
	return false
}

// withoutField returns the fields and their sizes without all fields with the
// given key.
func withoutField(fields []zapcore.Field, sizes []int, key string) ([]zapcore.Field, []int) {
	index := -1
	for i := range fields {
		if fields[i].Key == key {
			index = i
			break
		}
	}
	if index < 0 {
		return fields, sizes
	}

	filteredFields := make([]zapcore.Field, 0, len(fields)-1)
	filteredSizes := make([]int, 0, len(sizes)-1)
	for i := range fields {
		if fields[i].Key == key {
			continue
		}
		filteredFields = append(filteredFields, fields[i])
		filteredSizes = append(filteredSizes, sizes[i])
	}
	return filteredFields, filteredSizes
}

// hasNamespaceField reports whether the fields contain a namespace field.
func hasNamespaceField(fields []zapcore.Field) bool {
	for i := range fields {
		if fields[i].Type == zapcore.NamespaceType {
			return true
		}
	}
	return false
}

// stringValue returns the raw value of a string field: a string field, a byte
// string, an error or a stringer. The value of other fields is encoded as JSON
// and cannot be restored by a simple concatenation.
func stringValue(field zapcore.Field) (value string, ok bool) {
	if field.Type == zapcore.StringType {
		return field.String, true
	}
	if field.Type == zapcore.ByteStringType {
		bytes, isBytes := field.Interface.([]byte)
		if !isBytes {
			return "", false
		}
		return string(bytes), true
	}
	if field.Type == zapcore.ErrorType {
		err, isError := field.Interface.(error)
		if !isError || err == nil {
			return "", false
		}
		return errorValue(err)
	}
	if field.Type == zapcore.StringerType {
		stringer, isStringer := field.Interface.(fmt.Stringer)
		if !isStringer {
			return "", false
		}
		return stringerValue(stringer)
	}
	return "", false
}

// errorValue returns the text of an error. The Error method may panic, for
// example when the error is a typed nil pointer; such a field cannot be cut.
func errorValue(err error) (value string, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = "", false
		}
	}()
	return err.Error(), true
}

// stringerValue returns the text of a stringer. The String method may panic,
// for example when the value is a typed nil pointer; such a field cannot be
// cut.
func stringerValue(stringer fmt.Stringer) (value string, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = "", false
		}
	}()
	return stringer.String(), true
}

// escapedByteWidth maps a byte of a raw string to the number of bytes it takes
// in the JSON output: 1 for a regular byte, 2 for a character escaped with a
// backslash, 6 for a control character escaped as \u00XX and 1 for every byte of
// a valid multibyte UTF-8 sequence.
var escapedByteWidth = newEscapedByteWidth()

func newEscapedByteWidth() *[256]byte {
	widths := new([256]byte)
	for i := range widths {
		widths[i] = 1
	}
	for i := range controlByteBoundary {
		widths[i] = escapedControlWidth
	}
	widths['"'] = escapedSpecialWidth
	widths['\\'] = escapedSpecialWidth
	widths['\n'] = escapedSpecialWidth
	widths['\r'] = escapedSpecialWidth
	widths['\t'] = escapedSpecialWidth
	return widths
}

// escapedWidth returns the number of bytes s takes in the JSON output. s must be
// valid UTF-8.
func escapedWidth(s string) int {
	width := 0
	for i := range len(s) {
		width += int(escapedByteWidth[s[i]])
	}
	return width
}

// escapedBytesWidth returns the number of bytes s takes in the JSON output. s
// must be valid UTF-8.
func escapedBytesWidth(s []byte) int {
	width := 0
	for i := range s {
		width += int(escapedByteWidth[s[i]])
	}
	return width
}

// fieldKeyOverhead returns the number of bytes taken by the key of a field and
// its separating comma: ,"key":.
func fieldKeyOverhead(key string) int {
	return fieldKeySyntaxWidth + escapedWidth(key)
}

// chunkByBytes splits s into chunks whose JSON-escaped size does not exceed
// budget. It never splits a UTF-8 sequence. It returns false when a chunk
// cannot be built: s is not valid UTF-8 or a single character does not fit the
// budget.
func chunkByBytes(s string, budget int) ([]string, bool) {
	if budget <= 0 {
		return nil, false
	}
	if s == "" {
		return []string{""}, true
	}
	if !utf8.ValidString(s) {
		return nil, false
	}

	chunks := make([]string, 0, len(s)/budget+1)
	for len(s) > 0 {
		used := 0
		index := 0
		for index < len(s) {
			width := int(escapedByteWidth[s[index]])
			if used+width > budget {
				break
			}
			used += width
			index++
		}
		// Do not cut a UTF-8 sequence: step back to the beginning of the
		// current character.
		for index > 0 && index < len(s) && isUTF8Continuation(s[index]) {
			index--
		}
		if index == 0 {
			// The first character does not fit the budget.
			return nil, false
		}
		chunks = append(chunks, s[:index])
		s = s[index:]
	}
	return chunks, true
}

// isUTF8Continuation reports whether b is a continuation byte of a UTF-8
// sequence (10xxxxxx).
func isUTF8Continuation(b byte) bool {
	return b&utf8ContinuationMask == utf8ContinuationPrefix
}

// recordLen returns the size of the encoded record without the line ending.
func recordLen(line *buffer.Buffer) int {
	data := line.Bytes()
	size := len(data)
	if size > 0 && data[size-1] == '\n' {
		size--
	}
	if size > 0 && data[size-1] == '\r' {
		size--
	}
	return size
}
