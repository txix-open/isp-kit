package log_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"unicode/utf8"
	"uuid"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"go.uber.org/zap/zapcore"
)

const (
	testMaxLineKb    = 1
	testMaxLineBytes = testMaxLineKb * 1024

	envelopeLevelKey = "level"
	envelopeTimeKey  = "ts"
	envelopeMsgKey   = "msg"
)

func TestSplitDisabledByDefault(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{InitialLevel: log.InfoLevel})
	adapter.Info(context.Background(), "message", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Len(t, lines, 1)
	require.Greater(t, len(lines[0]), 1024)

	record := parseTestLine(t, lines[0])
	require.Equal(t, strings.Repeat("x", 4096), record["responseBody"])
	require.NotContains(t, record, log.SplitFieldPackageId)
}

func TestSplitSmallRecordEqualsUnsplit(t *testing.T) {
	t.Parallel()

	ctx := log.ToContext(context.Background(), log.String(requestid.LogKey, "req-42"))

	wholeAdapter, wholePath := newTestAdapter(t, log.Config{InitialLevel: log.InfoLevel})
	wholeAdapter.Info(ctx, "small message", log.String("responseBody", "short body"), log.Int("statusCode", 200))

	splitAdapter, splitPath := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	splitAdapter.Info(ctx, "small message", log.String("responseBody", "short body"), log.Int("statusCode", 200))

	wholeLines := readTestLines(t, wholePath)
	splitLines := readTestLines(t, splitPath)
	require.Len(t, wholeLines, 1)
	require.Len(t, splitLines, 1)

	wholeRecord := parseTestLine(t, wholeLines[0])
	splitRecord := parseTestLine(t, splitLines[0])
	delete(wholeRecord, envelopeTimeKey)
	delete(splitRecord, envelopeTimeKey)
	require.Equal(t, wholeRecord, splitRecord)
}

func TestSplitOversizedField(t *testing.T) {
	t.Parallel()

	ctx := log.ToContext(context.Background(), log.String(requestid.LogKey, "req-42"))
	responseBody := `{"added":[],"changed":[],"removed":[` + strings.Repeat(`{"block":"ADDRESS"},`, 300) + `]}`

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	adapter.Info(ctx, "http handler: response",
		log.Int("statusCode", 200),
		log.Int64("elapsedTimeMs", 302),
		log.String("responseBody", responseBody),
	)

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)

	var (
		packageId string
		startedAt string
	)
	for i, line := range lines {
		require.LessOrEqual(t, len(line), testMaxLineBytes)
		require.True(t, utf8.ValidString(line))

		record := parseTestLine(t, line)
		require.Equal(t, "info", record[envelopeLevelKey])
		require.Equal(t, "req-42", record[requestid.LogKey])
		require.Equal(t, 1, strings.Count(line, `"`+requestid.LogKey+`"`), "requestId must be in the root and only once")

		require.EqualValues(t, i+1, record[log.SplitFieldPackageIndex])
		require.EqualValues(t, len(lines), record[log.SplitFieldPackageTotal])

		currentPackageId := recordString(t, record, log.SplitFieldPackageId)
		_, err := uuid.Parse(currentPackageId)
		require.NoError(t, err)

		currentStartedAt := recordString(t, record, envelopeTimeKey)
		require.Equal(t, "http handler: response", record[envelopeMsgKey])
		if i == 0 {
			packageId = currentPackageId
			startedAt = currentStartedAt
			// Small fields are packed before the large one.
			require.NotContains(t, record, "responseBody")
			require.EqualValues(t, 200, record["statusCode"])
			require.EqualValues(t, 302, record["elapsedTimeMs"])
			continue
		}
		require.Equal(t, packageId, currentPackageId)
		require.Equal(t, startedAt, currentStartedAt)
		require.Contains(t, record, "responseBody")
	}

	require.Equal(t, responseBody, collectResponseBody(t, lines))
}

// TestSplitOversizedByteString checks that an oversized byte string field is
// split and the parts restore the original bytes.
func TestSplitOversizedByteString(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	body := []byte(strings.Repeat("response ", 500))
	adapter.Info(context.Background(), "http handler: response",
		log.ByteString("responseBody", body),
	)

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	requirePartsFitLimit(t, lines)
	require.Equal(t, string(body), collectResponseBody(t, lines))
}

// adapterStringer implements fmt.Stringer: logged via log.Any it becomes a
// zapcore.StringerType field.
type adapterStringer string

func (s adapterStringer) String() string {
	return string(s)
}

// TestSplitOversizedStringer checks that an oversized fmt.Stringer field is
// split and the parts restore the original string.
func TestSplitOversizedStringer(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	value := strings.Repeat("значение ", 500)
	adapter.Info(context.Background(), "http handler: response",
		log.Any("responseBody", adapterStringer(value)),
	)

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	requirePartsFitLimit(t, lines)
	require.Equal(t, value, collectResponseBody(t, lines))
}

func TestSplitKeepsMessageInEveryPart(t *testing.T) {
	t.Parallel()

	ctx := log.ToContext(context.Background(), log.String(requestid.LogKey, "req-42"))
	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	adapter.Info(ctx, "http handler: response", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	requirePartsFitLimit(t, lines)

	var startedAt string
	for i, line := range lines {
		record := parseTestLine(t, line)
		require.Equal(t, "info", record[envelopeLevelKey])
		require.Equal(t, "req-42", record[requestid.LogKey])
		require.EqualValues(t, i+1, record[log.SplitFieldPackageIndex])

		require.Equal(t, "http handler: response", record[envelopeMsgKey])
		if i == 0 {
			startedAt = recordString(t, record, envelopeTimeKey)
			continue
		}
		require.Equal(t, startedAt, recordString(t, record, envelopeTimeKey))
	}
}

func TestSplitAssemblyRestoresRecord(t *testing.T) {
	t.Parallel()

	ctx := log.ToContext(context.Background(), log.String(requestid.LogKey, "req-42"))
	responseBody := strings.Repeat(`{"block":"ADDRESS","value":"улица Ленина, д. 1"},`, 200)
	fields := []log.Field{
		log.Int("statusCode", 200),
		log.Int64("elapsedTimeMs", 302),
		log.String("responseBody", responseBody),
	}

	wholeAdapter, wholePath := newTestAdapter(t, log.Config{InitialLevel: log.InfoLevel})
	wholeAdapter.Info(ctx, "http handler: response", fields...)

	splitAdapter, splitPath := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	splitAdapter.Info(ctx, "http handler: response", fields...)

	wholeLines := readTestLines(t, wholePath)
	splitLines := readTestLines(t, splitPath)
	require.Len(t, wholeLines, 1)
	require.Greater(t, len(splitLines), 1)
	requirePartsFitLimit(t, splitLines)

	assembled := assembleRecord(t, splitLines)
	delete(assembled, envelopeTimeKey)
	original := parseTestLine(t, wholeLines[0])
	delete(original, envelopeTimeKey)
	require.Equal(t, original, assembled)
}

func TestSplitPartsFitLimit(t *testing.T) {
	t.Parallel()

	// A deterministic property test over an alphabet that exercises escaping,
	// multi-byte runes and control characters.
	alphabet := []rune("abcXYZ019 \"\\\n\t\r\x01яю😀{}[],:")

	for caseIndex := range 20 {
		var value strings.Builder
		for position := 0; value.Len() < 6000; position++ {
			value.WriteRune(alphabet[(position*7+caseIndex*13)%len(alphabet)])
		}

		adapter, path := newTestAdapter(t, log.Config{
			InitialLevel: log.InfoLevel,
			Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
		})
		adapter.Info(context.Background(), "message",
			log.String("responseBody", value.String()),
			log.Int("statusCode", 200),
		)

		lines := readTestLines(t, path)
		require.Greater(t, len(lines), 1)
		requirePartsFitLimit(t, lines)

		for _, line := range lines {
			require.True(t, utf8.ValidString(line))
		}
		require.Equal(t, value.String(), collectResponseBody(t, lines))
	}
}

func TestSplitManySmallFields(t *testing.T) {
	t.Parallel()

	fields := make([]log.Field, 0, 200)
	for i := range 200 {
		fields = append(fields, log.String(fmt.Sprintf("field%03d", i), strings.Repeat("v", 40)))
	}

	wholeAdapter, wholePath := newTestAdapter(t, log.Config{InitialLevel: log.InfoLevel})
	wholeAdapter.Info(context.Background(), "message", fields...)

	splitAdapter, splitPath := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	splitAdapter.Info(context.Background(), "message", fields...)

	wholeLines := readTestLines(t, wholePath)
	splitLines := readTestLines(t, splitPath)
	require.Len(t, wholeLines, 1)
	require.Greater(t, len(splitLines), 1)
	requirePartsFitLimit(t, splitLines)

	assembled := assembleRecord(t, splitLines)
	delete(assembled, envelopeTimeKey)
	original := parseTestLine(t, wholeLines[0])
	delete(original, envelopeTimeKey)
	require.Equal(t, original, assembled)

	// Small fields are packed, not cut: every field occurs exactly once.
	for i := range 200 {
		key := fmt.Sprintf("field%03d", i)
		count := 0
		for _, line := range splitLines {
			if strings.Contains(line, `"`+key+`"`) {
				count++
			}
		}
		require.Equal(t, 1, count, "field %q must be packed into one part", key)
	}
}

func TestSplitMixOfSmallAndLargeFields(t *testing.T) {
	t.Parallel()

	responseBody := strings.Repeat(`{"block":"ADDRESS"},`, 300)
	splitAdapter, splitPath := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	splitAdapter.Info(context.Background(), "message",
		log.String("before", "first"),
		log.String("responseBody", responseBody),
		log.String("after", "last"),
	)

	lines := readTestLines(t, splitPath)
	require.Greater(t, len(lines), 2)
	requirePartsFitLimit(t, lines)

	// The field order is preserved: fields before the large one are packed into
	// the first part, fields after it go to the last part.
	require.Contains(t, lines[0], `"before"`)
	require.Contains(t, lines[len(lines)-1], `"after"`)
	require.NotContains(t, lines[0], `"responseBody"`)

	require.Equal(t, responseBody, collectResponseBody(t, lines))
	require.Equal(t, "first", parseTestLine(t, lines[0])["before"])
	require.Equal(t, "last", parseTestLine(t, lines[len(lines)-1])["after"])
}

func TestSplitKeepsSamplingDisabledByDefault(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})

	const records = 150
	for range records {
		adapter.Info(context.Background(), "message", log.Int("index", 1))
	}

	lines := readTestLines(t, path)
	require.Len(t, lines, records)
}

func TestSplitDisabledInDevMode(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		IsInDevMode:  true,
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	adapter.Info(context.Background(), "message", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Len(t, lines, 1)
	require.Greater(t, len(lines[0]), testMaxLineBytes)
	require.NotContains(t, lines[0], log.SplitFieldPackageId)
}

func TestSplitInactiveWithoutLimit(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true},
	})
	adapter.Info(context.Background(), "message", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Len(t, lines, 1)
	require.Greater(t, len(lines[0]), testMaxLineBytes)
	require.NotContains(t, lines[0], log.SplitFieldPackageId)
}

func TestSplitWithoutRequestId(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	adapter.Info(context.Background(), "message", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	for _, line := range lines {
		require.NotContains(t, parseTestLine(t, line), requestid.LogKey)
	}
}

func TestSplitSamplingDecisionIsMadeOncePerRecord(t *testing.T) {
	t.Parallel()

	responseBody := strings.Repeat("x", 4096)
	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Sampling:     &log.SamplingConfig{Initial: 1, Thereafter: 0},
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})

	for range 3 {
		adapter.Info(context.Background(), "message", log.String("responseBody", responseBody))
	}

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	requirePartsFitLimit(t, lines)

	var packageId string
	for i, line := range lines {
		record := parseTestLine(t, line)
		require.EqualValues(t, i+1, record[log.SplitFieldPackageIndex])
		require.EqualValues(t, len(lines), record[log.SplitFieldPackageTotal])

		current := recordString(t, record, log.SplitFieldPackageId)
		if i == 0 {
			packageId = current
		}
		require.Equal(t, packageId, current)
	}
}

func TestSplitHooksCalledOncePerRecord(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
		Hooks: []func(entry zapcore.Entry) error{
			func(zapcore.Entry) error {
				calls.Add(1)
				return nil
			},
		},
	})
	adapter.Info(context.Background(), "message", log.String("responseBody", strings.Repeat("x", 4096)))

	lines := readTestLines(t, path)
	require.Greater(t, len(lines), 1)
	require.EqualValues(t, 1, calls.Load())
}

func TestSplitLimitSmallerThanEnvelope(t *testing.T) {
	t.Parallel()

	adapter, path := newTestAdapter(t, log.Config{
		InitialLevel: log.InfoLevel,
		Split:        &log.SplitConfig{Enabled: true, MaxLineKb: testMaxLineKb},
	})
	adapter.Info(context.Background(), strings.Repeat("m", 4096), log.String("key", "value"))

	lines := readTestLines(t, path)
	require.Len(t, lines, 1)
	require.Greater(t, len(lines[0]), testMaxLineBytes)
	require.NotContains(t, parseTestLine(t, lines[0]), log.SplitFieldPackageId)
}

// assembleRecord restores the original record from its parts as described in
// the package documentation: group by logPackageId, sort by logPackageIndex,
// concatenate the values of a field that occurs more than once and keep the
// value of a field that occurs once. The level, time, message and request id
// are the same in every part.
func assembleRecord(t *testing.T, lines []string) map[string]any {
	t.Helper()

	type part struct {
		index  int
		record map[string]any
	}

	parts := make([]part, 0, len(lines))
	for _, line := range lines {
		record := parseTestLine(t, line)
		index, ok := record[log.SplitFieldPackageIndex].(float64)
		require.True(t, ok, "part must have a numeric %q", log.SplitFieldPackageIndex)
		parts = append(parts, part{index: int(index), record: record})
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i].index < parts[j].index
	})

	for i, current := range parts {
		require.Equal(t, i+1, current.index)
		require.EqualValues(t, len(parts), current.record[log.SplitFieldPackageTotal])
		if i > 0 {
			require.Equal(t,
				recordString(t, parts[0].record, log.SplitFieldPackageId),
				recordString(t, current.record, log.SplitFieldPackageId),
			)
		}
	}

	result := make(map[string]any, len(parts[0].record))
	for _, key := range []string{envelopeLevelKey, envelopeTimeKey, envelopeMsgKey, requestid.LogKey} {
		if value, ok := parts[0].record[key]; ok {
			result[key] = value
		}
	}

	values := make(map[string][]any, len(parts[0].record))
	for _, current := range parts {
		for key, value := range current.record {
			switch key {
			case envelopeLevelKey, envelopeTimeKey, envelopeMsgKey,
				log.SplitFieldPackageId, log.SplitFieldPackageIndex, log.SplitFieldPackageTotal,
				requestid.LogKey:
				continue
			}
			values[key] = append(values[key], value)
		}
	}

	for key, list := range values {
		if len(list) == 1 {
			result[key] = list[0]
			continue
		}

		var joined strings.Builder
		for _, value := range list {
			text, ok := value.(string)
			require.True(t, ok, "repeated field %q must be a string", key)
			joined.WriteString(text)
		}
		result[key] = joined.String()
	}
	return result
}

func requirePartsFitLimit(t *testing.T, lines []string) {
	t.Helper()

	for _, line := range lines {
		require.LessOrEqual(t, len(line), testMaxLineBytes, line)
	}
}

// collectResponseBody joins the responseBody field value of the parts in the
// order they were written.
func collectResponseBody(t *testing.T, lines []string) string {
	t.Helper()

	var collected strings.Builder
	for _, line := range lines {
		value, ok := parseTestLine(t, line)["responseBody"]
		if !ok {
			continue
		}
		text, ok := value.(string)
		require.True(t, ok, "field %q must be a string", "responseBody")
		collected.WriteString(text)
	}
	return collected.String()
}

func newTestAdapter(t *testing.T, cfg log.Config) (*log.Adapter, string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "log.json")
	cfg.DisableDefaultOutput = true
	cfg.OutputPaths = []string{path}

	adapter, err := log.NewFromConfig(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, adapter.Sync())
	})
	return adapter, path
}

func readTestLines(t *testing.T, path string) []string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func parseTestLine(t *testing.T, line string) map[string]any {
	t.Helper()

	var record map[string]any
	require.NoError(t, json.Unmarshal([]byte(line), &record), line)
	return record
}

func recordString(t *testing.T, record map[string]any, key string) string {
	t.Helper()

	value, ok := record[key].(string)
	require.True(t, ok, "field %q must be a string", key)
	return value
}
