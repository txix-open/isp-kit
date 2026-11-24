package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDefaultLimits(t *testing.T) {
	t.Parallel()

	require.Equal(t, 218, DefaultMaxLineKb)
	require.Equal(t, 223232, DefaultMaxLineBytes)
}

func TestEscapedByteWidth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		value byte
		want  byte
	}{
		{name: "ascii", value: 'a', want: 1},
		{name: "space", value: ' ', want: 1},
		{name: "quote", value: '"', want: 2},
		{name: "backslash", value: '\\', want: 2},
		{name: "newline", value: '\n', want: 2},
		{name: "carriage return", value: '\r', want: 2},
		{name: "tab", value: '\t', want: 2},
		{name: "control", value: 0x01, want: 6},
		{name: "delete", value: 0x7f, want: 1},
		{name: "utf8 continuation", value: 0x80, want: 1},
		{name: "utf8 lead", value: 0xff, want: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testCase.want, escapedByteWidth[testCase.value])
		})
	}
}

func TestChunkByBytes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		value  string
		budget int
	}{
		{name: "ascii", value: strings.Repeat("abc", 100), budget: 32},
		{name: "cyrillic", value: strings.Repeat("привет", 100), budget: 33},
		{name: "quotes", value: strings.Repeat(`a"b`, 100), budget: 17},
		{name: "backslashes", value: strings.Repeat(`a\b`, 100), budget: 13},
		{name: "controls", value: strings.Repeat("a\x01b", 100), budget: 11},
		{name: "emoji", value: strings.Repeat("😀!", 100), budget: 21},
		{name: "newlines", value: strings.Repeat("a\nb", 100), budget: 9},
		{name: "mixed", value: strings.Repeat("я\"\\\x02😀", 100), budget: 27},
		{name: "single character budget", value: "😀", budget: 4},
		{name: "budget bigger than the value", value: "short", budget: 1024},
		{name: "empty value", value: "", budget: 16},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			chunks, ok := chunkByBytes(testCase.value, testCase.budget)
			require.True(t, ok)
			require.Equal(t, testCase.value, strings.Join(chunks, ""))

			for _, chunk := range chunks {
				require.True(t, utf8.ValidString(chunk), "chunk must not split a rune")
				require.LessOrEqual(t, escapedWidth(chunk), testCase.budget)
			}

			if testCase.value == "" {
				require.Equal(t, []string{""}, chunks)
			}
		})
	}
}

func TestChunkByBytesRejects(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		value  string
		budget int
	}{
		{name: "zero budget", value: "abc", budget: 0},
		{name: "negative budget", value: "abc", budget: -1},
		{name: "invalid UTF-8", value: "abc\xffdef", budget: 1024},
		{name: "rune does not fit", value: "😀", budget: 3},
		{name: "cyrillic rune does not fit", value: "я", budget: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			chunks, ok := chunkByBytes(testCase.value, testCase.budget)
			require.False(t, ok)
			require.Nil(t, chunks)
		})
	}
}

func TestFieldSizesAreAdditive(t *testing.T) {
	t.Parallel()

	encConfig := testEncoderConfig()
	encoder := zapcore.NewJSONEncoder(encConfig)
	bareEncoder := zapcore.NewJSONEncoder(fieldOnlyEncoderConfig(encConfig))
	entry := testEntry()
	fields := []zapcore.Field{
		zap.String("string", "value"),
		zap.Int("int", 42),
		zap.Int("negative", -12345),
		zap.Bool("bool", true),
		zap.Bool("false", false),
		zap.Uint64("uint", 18446744073709551615),
		zap.ByteString("bytes", []byte("данные")),
		zap.String("special", "quote \" slash \\ newline \n ctrl \x01 tab \t"),
		zap.Any("object", map[string]any{"x": 1, "y": []int{1, 2}}),
		zap.Error(errTestCause),
		zap.Skip(),
		zap.String("invalid", "a\xffb"),
		zap.ByteString("invalidBytes", []byte{0xff, 0xfe}),
		zap.String("key\xff", "value"),
	}

	base := encodedLen(t, encoder, entry, nil)
	full := encodedLen(t, encoder, entry, fields)

	total := 0
	for _, field := range fields {
		size, err := fieldSize(bareEncoder, field)
		require.NoError(t, err)
		total += size

		if cheap, ok := cheapFieldSize(field); ok {
			require.Equal(t, size, cheap, "field %q", field.Key)
		}
	}
	require.Equal(t, full-base, total)
}

func TestFieldSizeOfNamespace(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(DefaultMaxLineKb)
	sizer := &fieldSizer{encConfig: core.bareEncConfig, limit: core.maxLineBytes}
	_, err := sizer.size(zap.Namespace("namespace"))
	require.ErrorIs(t, err, errCannotSplit)
}

func TestCheapFieldSizeRejectsInvalidKey(t *testing.T) {
	t.Parallel()

	_, ok := cheapFieldSize(zap.String("key\xff", "value"))
	require.False(t, ok)
}

func TestPlanPartsPacksFieldsInOrder(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	entry := testEntry()

	fields := make([]zapcore.Field, 0, 41)
	fields = append(fields, zap.Skip())
	for i := range 40 {
		fields = append(fields, zap.String(fmt.Sprintf("field%02d", i), strings.Repeat("v", 60)))
	}

	parts := core.planParts(entry, fields, testFieldSizes(t, core, fields), "package-id", zapcore.Field{}, false)
	require.Greater(t, len(parts), 1)

	packed := make([]string, 0, len(fields))
	for i, part := range parts {
		require.NotEmpty(t, part)
		for _, field := range part {
			packed = append(packed, field.Key)
		}

		partFields := append(metaFields("package-id", i+1, len(parts), zapcore.Field{}, false), part...)
		require.LessOrEqual(t, partLineLen(t, core, entry, partFields), core.maxLineBytes)
	}

	expected := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.Type == zapcore.SkipType {
			continue
		}
		expected = append(expected, field.Key)
	}
	require.Equal(t, expected, packed)
}

func TestPlanPartsChunksLargeField(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	entry := testEntry()
	value := strings.Repeat("данные ", 500)

	fields := []zapcore.Field{
		zap.String("statusCode", "200"),
		zap.String("responseBody", value),
	}
	parts := core.planParts(entry, fields, testFieldSizes(t, core, fields), "package-id", zapcore.Field{}, false)
	require.Greater(t, len(parts), 2)

	// The field before the large one is packed into its own part.
	require.Len(t, parts[0], 1)
	require.Equal(t, "statusCode", parts[0][0].Key)

	// Every chunk keeps the original field name and the parts are ordered.
	var joined strings.Builder
	for i, part := range parts[1:] {
		require.Len(t, part, 1)
		require.Equal(t, "responseBody", part[0].Key)
		require.Equal(t, zapcore.StringType, part[0].Type)
		joined.WriteString(part[0].String)

		partFields := append(metaFields("package-id", i+2, len(parts), zapcore.Field{}, false), part...)
		require.LessOrEqual(t, partLineLen(t, core, entry, partFields), core.maxLineBytes)
	}
	require.Equal(t, value, joined.String())
}

func TestPlanPartsRejectsNonStringLargeField(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	fields := []zapcore.Field{
		zap.Any("object", map[string]string{"key": strings.Repeat("x", 4096)}),
	}

	parts := core.planParts(testEntry(), fields, testFieldSizes(t, core, fields), "package-id", zapcore.Field{}, false)
	require.Nil(t, parts)
}

func TestPlanPartsRejectsOversizedEnvelope(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	entry := testEntry()
	entry.Message = strings.Repeat("m", 4096)

	fields := []zapcore.Field{zap.String("key", "value")}
	parts := core.planParts(entry, fields, testFieldSizes(t, core, fields), "package-id", zapcore.Field{}, false)
	require.Nil(t, parts)
}

func TestChunkFieldRejectsErrorWithCauses(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	field := zap.Error(causesError{
		text:   strings.Repeat("e", 4096),
		causes: []error{errTestCause},
	})

	size := testFieldSize(t, core, field)
	require.Nil(t, core.chunkField(field, size, core.maxLineBytes), "an error with causes must not lose them")
}

func TestChunkFieldRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	field := zap.String("responseBody", strings.Repeat("\xff", 4096))

	size := testFieldSize(t, core, field)
	require.Nil(t, core.chunkField(field, size, core.maxLineBytes))
}

func TestSplittingCoreWritesSmallRecordAsIs(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger.Info("small message", zap.String("key", "value"))

	lines := recorder.collected()
	require.Len(t, lines, 1)

	record := parseInternalLine(t, lines[0])
	require.Equal(t, "small message", record["msg"])
	require.Equal(t, "value", record["key"])
	require.NotContains(t, record, SplitFieldPackageId)
}

func TestSplittingCoreWritesParts(t *testing.T) {
	t.Parallel()

	const maxLineKb = 1
	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: maxLineKb})
	responseBody := strings.Repeat(`{"block":"ADDRESS"},`, 200)

	logger.Info("http handler: response",
		zap.String(requestid.LogKey, "req-42"),
		zap.Int("statusCode", 200),
		zap.String("responseBody", responseBody),
	)

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)

	var (
		packageId string
		startedAt string
	)
	for i, line := range lines {
		require.LessOrEqual(t, len(line), maxLineKb*1024)
		require.Equal(t, 1, strings.Count(line, `"`+requestid.LogKey+`"`))

		record := parseInternalLine(t, line)
		require.Equal(t, "info", record["level"])
		require.Equal(t, "http handler: response", record["msg"])
		require.Equal(t, "req-42", record["requestId"])
		require.EqualValues(t, i+1, record[SplitFieldPackageIndex])
		require.EqualValues(t, len(lines), record[SplitFieldPackageTotal])

		if i == 0 {
			packageId = recordStringInternal(t, record, SplitFieldPackageId)
			startedAt = recordStringInternal(t, record, "ts")
			require.NotContains(t, record, "responseBody")
			continue
		}
		require.Contains(t, record, "responseBody")
		require.Equal(t, packageId, recordStringInternal(t, record, SplitFieldPackageId))
		require.Equal(t, startedAt, recordStringInternal(t, record, "ts"))
	}
}

func TestSplittingCoreKeepsLoggerFields(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger = logger.With(zap.String("component", "worker"))
	logger.Info("message", zap.String("responseBody", strings.Repeat("x", 4096)))

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)
	for _, line := range lines {
		require.Equal(t, "worker", parseInternalLine(t, line)["component"])
	}
}

func TestSplittingCoreKeepsRequestIdFromLoggerFields(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger = logger.With(zap.String(requestid.LogKey, "req-7"))
	logger.Info("message", zap.String("responseBody", strings.Repeat("x", 4096)))

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)
	for _, line := range lines {
		require.Equal(t, "req-7", parseInternalLine(t, line)["requestId"])
		require.Equal(t, 1, strings.Count(line, `"`+requestid.LogKey+`"`))
	}
}

func TestSplittingCoreWritesOversizedEnvelopeAsIs(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger.Info(strings.Repeat("m", 4096), zap.String("key", "value"))

	lines := recorder.collected()
	require.Len(t, lines, 1)
	require.Greater(t, len(lines[0]), 1024)
	require.NotContains(t, parseInternalLine(t, lines[0]), SplitFieldPackageId)
}

func TestSplittingCoreNamespaceFallback(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger.Info("message",
		zap.Namespace("namespace"),
		zap.String("responseBody", strings.Repeat("x", 4096)),
	)

	lines := recorder.collected()
	require.Len(t, lines, 1)
	require.NotContains(t, parseInternalLine(t, lines[0]), SplitFieldPackageId)
}

func TestSplittingCoreInvalidUTF8Fallback(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger.Info("message", zap.String("responseBody", strings.Repeat("\xff", 4096)))

	lines := recorder.collected()
	require.Len(t, lines, 1)
	require.NotContains(t, parseInternalLine(t, lines[0]), SplitFieldPackageId)
}

func TestSplittingCoreInvalidMessageStillSplitsField(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger.Info("message\xff", zap.String("responseBody", strings.Repeat("x", 4096)))

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)

	var joined strings.Builder
	for _, line := range lines {
		require.LessOrEqual(t, len(line), 1024)
		record := parseInternalLine(t, line)
		require.Contains(t, record, "msg")
		value, ok := record["responseBody"].(string)
		if ok {
			joined.WriteString(value)
		}
	}
	require.Equal(t, strings.Repeat("x", 4096), joined.String())
}

func TestSplittingCoreReturnsSinkError(t *testing.T) {
	t.Parallel()

	encConfig := testEncoderConfig()
	recorder := &memoryRecorder{writeErr: errTestSink}
	core := newSplittingCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.Lock(recorder),
		zapcore.InfoLevel,
		encConfig,
		SplitConfig{Enabled: true, MaxLineKb: 1},
	)

	err := core.Write(testEntry(), []zapcore.Field{
		zap.String("responseBody", strings.Repeat("x", 4096)),
	})
	require.ErrorIs(t, err, errTestSink)
}

// TestSplitSamplerAndHooksAtErrorLevel protects the core order split core ->
// sampler -> hooks: the sampler decision and the hooks must be applied to the
// original record once, at every level, not only at info level.
func TestSplitSamplerAndHooksAtErrorLevel(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger = logger.WithOptions(samplerOption(&SamplingConfig{Initial: 1, Thereafter: 0}))
	var calls atomic.Int32
	logger = logger.WithOptions(zap.Hooks(func(_ zapcore.Entry) error {
		calls.Add(1)
		return nil
	}))

	for range 5 {
		logger.Error("boom", zap.String("responseBody", strings.Repeat("x", 4096)))
	}

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)
	require.Len(t, uniquePackageIds(t, lines), 1)
	require.EqualValues(t, 1, calls.Load())
}

func TestSplitConcurrentWrites(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for range 50 {
				logger.Info("concurrent message",
					zap.String("responseBody", strings.Repeat("x", 3000)),
					zap.Int("worker", worker),
				)
			}
		}(i)
	}
	wg.Wait()

	require.Len(t, uniquePackageIds(t, recorder.collected()), 400)
}

// TestSplittingCoreRequestIdFromLoggerAndCall checks that a request id attached
// by With and a request id in the record are not written twice: the call field
// is removed and the logger field is used.
func TestSplittingCoreRequestIdFromLoggerAndCall(t *testing.T) {
	t.Parallel()

	logger, recorder := newTestLogger(SplitConfig{Enabled: true, MaxLineKb: 1})
	logger = logger.With(zap.String(requestid.LogKey, "req-7"))
	logger.Info("message",
		zap.String(requestid.LogKey, "req-42"),
		zap.String("responseBody", strings.Repeat("x", 4096)),
	)

	lines := recorder.collected()
	require.Greater(t, len(lines), 1)
	for _, line := range lines {
		require.Equal(t, "req-7", parseInternalLine(t, line)[requestid.LogKey])
		require.Equal(t, 1, strings.Count(line, `"`+requestid.LogKey+`"`))
	}
}

// TestChunkFieldChunksPlainError covers the error chunking success path: an
// error without causes is cut like a string.
func TestChunkFieldChunksPlainError(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	value := strings.Repeat("ошибка ", 500)
	field := zap.Error(plainError(value))

	size := testFieldSize(t, core, field)
	chunks := core.chunkField(field, size, core.maxLineBytes)
	require.NotEmpty(t, chunks)

	var joined strings.Builder
	for _, chunk := range chunks {
		require.Len(t, chunk, 1)
		require.Equal(t, field.Key, chunk[0].Key)
		joined.WriteString(chunk[0].String)
	}
	require.Equal(t, value, joined.String())
}

func uniquePackageIds(t *testing.T, lines []string) []string {
	t.Helper()

	ids := make(map[string]struct{}, len(lines))
	for _, line := range lines {
		record := parseInternalLine(t, line)
		value, ok := record[SplitFieldPackageId]
		require.True(t, ok, "part must have %q", SplitFieldPackageId)
		id, isString := value.(string)
		require.True(t, isString, "%q must be a string", SplitFieldPackageId)
		ids[id] = struct{}{}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	return result
}

func TestChunkFieldChunksStringAndByteString(t *testing.T) {
	t.Parallel()

	core := newTestSplitCore(1)
	value := strings.Repeat("значение \"quoted\" ", 500)

	for _, field := range []zapcore.Field{
		zap.String("string", value),
		zap.ByteString("bytes", []byte(value)),
	} {
		size := testFieldSize(t, core, field)
		chunks := core.chunkField(field, size, core.maxLineBytes)
		require.NotEmpty(t, chunks)

		var joined strings.Builder
		for _, chunk := range chunks {
			require.Len(t, chunk, 1)
			require.Equal(t, field.Key, chunk[0].Key)
			require.Equal(t, zapcore.StringType, chunk[0].Type)
			joined.WriteString(chunk[0].String)
		}
		require.Equal(t, value, joined.String())
	}
}

func testEntry() zapcore.Entry {
	return zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
		Message: "http handler: response",
	}
}

func testEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg
}

func encodedLen(t *testing.T, encoder zapcore.Encoder, entry zapcore.Entry, fields []zapcore.Field) int {
	t.Helper()

	line, err := encoder.EncodeEntry(entry, fields)
	require.NoError(t, err)
	defer line.Free()
	return recordLen(line)
}

func partLineLen(
	t *testing.T,
	core *splittingCore,
	entry zapcore.Entry,
	fields []zapcore.Field,
) int {
	t.Helper()

	return encodedLen(t, core.encoder, entry, fields)
}

func testFieldSizes(t *testing.T, core *splittingCore, fields []zapcore.Field) []int {
	t.Helper()

	sizer := &fieldSizer{encConfig: core.bareEncConfig, limit: core.maxLineBytes}
	sizes := make([]int, len(fields))
	for i := range fields {
		size, err := sizer.size(fields[i])
		require.NoError(t, err)
		sizes[i] = size
	}
	return sizes
}

func testFieldSize(t *testing.T, core *splittingCore, field zapcore.Field) int {
	t.Helper()

	return testFieldSizes(t, core, []zapcore.Field{field})[0]
}

func newTestSplitCore(maxLineKb int) *splittingCore {
	encConfig := testEncoderConfig()
	return newSplittingCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.Lock(zapcore.AddSync(io.Discard)),
		zapcore.InfoLevel,
		encConfig,
		SplitConfig{Enabled: true, MaxLineKb: maxLineKb},
	)
}

func newTestLogger(cfg SplitConfig) (*zap.Logger, *memoryRecorder) {
	encConfig := testEncoderConfig()
	recorder := &memoryRecorder{}
	core := newSplittingCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.Lock(recorder),
		zapcore.InfoLevel,
		encConfig,
		cfg,
	)
	return zap.New(core), recorder
}

func parseInternalLine(t *testing.T, line string) map[string]any {
	t.Helper()

	var record map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &record), line)
	return record
}

func recordStringInternal(t *testing.T, record map[string]any, key string) string {
	t.Helper()

	value, ok := record[key].(string)
	require.True(t, ok, "field %q must be a string", key)
	return value
}

var (
	errTestCause = errors.New("cause")
	errTestSink  = errors.New("sink is broken")
)

// plainError is a plain error text without causes: zap encodes it as a single
// string field.
type plainError string

func (e plainError) Error() string {
	return string(e)
}

type causesError struct {
	text   string
	causes []error
}

func (e causesError) Error() string {
	return e.text
}

func (e causesError) Errors() []error {
	return e.causes
}

// memoryRecorder collects the lines written by the splitting core.
type memoryRecorder struct {
	mu       sync.Mutex
	data     bytes.Buffer
	writeErr error
}

func (r *memoryRecorder) Write(p []byte) (int, error) {
	if r.writeErr != nil {
		return 0, r.writeErr
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data.Write(p)
}

func (r *memoryRecorder) Sync() error {
	return nil
}

func (r *memoryRecorder) collected() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	text := strings.TrimSuffix(r.data.String(), "\n")
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}

func newBenchmarkLogger() *zap.Logger {
	encConfig := testEncoderConfig()
	core := newSplittingCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.Lock(zapcore.AddSync(io.Discard)),
		zapcore.InfoLevel,
		encConfig,
		SplitConfig{Enabled: true, MaxLineKb: DefaultMaxLineKb},
	)
	return zap.New(core)
}

func BenchmarkWriteSmall(b *testing.B) {
	logger := newBenchmarkLogger()
	field := zap.String("key", "value")

	b.ReportAllocs()
	for b.Loop() {
		logger.Info("small message", field)
	}
}

func BenchmarkWriteNoSplit(b *testing.B) {
	encConfig := testEncoderConfig()
	baseCore := zapcore.NewCore(zapcore.NewJSONEncoder(encConfig), zapcore.AddSync(io.Discard), zapcore.InfoLevel)
	logger := zap.New(baseCore)
	field := zap.String("responseBody", strings.Repeat("x", 1<<20))

	b.ReportAllocs()
	for b.Loop() {
		logger.Info("message", field)
	}
}

func BenchmarkWriteSplit1MB(b *testing.B) {
	logger := newBenchmarkLogger()
	field := zap.String("responseBody", strings.Repeat("x", 1<<20))

	b.ReportAllocs()
	for b.Loop() {
		logger.Info("message", field)
	}
}

func BenchmarkWriteFit200KB(b *testing.B) {
	logger := newBenchmarkLogger()
	field := zap.String("responseBody", strings.Repeat("x", 200<<10))

	b.ReportAllocs()
	for b.Loop() {
		logger.Info("message", field)
	}
}

func BenchmarkWriteManyFields(b *testing.B) {
	logger := newBenchmarkLogger()
	fields := make([]zapcore.Field, 0, 100)
	for i := range 100 {
		fields = append(fields, zap.String(fmt.Sprintf("field%03d", i), "value"))
	}

	b.ReportAllocs()
	for b.Loop() {
		logger.Info("message", fields...)
	}
}
