package file_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/file"
	"github.com/txix-open/isp-kit/test"
	"gopkg.in/natefinch/lumberjack.v2"
)

const partFile = "app.log"

var backupNamePattern = regexp.MustCompile(`^app-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}\.\d{3}(-\d+)?\.log$`)

type rotatingWriterSuite struct {
	suite.Suite

	test *test.Test

	dir  string
	name string
}

func (t *rotatingWriterSuite) SetupTest() {
	t.test, _ = test.New(t.T())
	t.dir = t.test.T().TempDir()
	t.name = filepath.Join(t.dir, partFile)
}

func TestRotatingWriter(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(rotatingWriterSuite))
}

func (t *rotatingWriterSuite) Test_NewFileWriter_LumberjackWithoutPartSizeKb() {
	w := file.NewFileWriter(file.Output{File: t.name, MaxSizeMb: 512})
	t.Require().IsType(&lumberjack.Logger{}, w)

	writeAll(t.T(), w, makePayload('a', 100))
	t.Require().NoError(w.Close())
	t.Require().Contains(string(readPart(t.T(), t.name)), "a")

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(plain)
	t.Require().Empty(gz)
}

func (t *rotatingWriterSuite) Test_NewFileWriter_RotatingWriterWithPartSizeKb() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})

	_, isLumberjack := w.(*lumberjack.Logger)
	t.Require().False(isLumberjack)
}

func (t *rotatingWriterSuite) Test_Rotation_AtPartSize() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	first := makePayload('a', 3*1024)
	second := makePayload('b', 3*1024)
	writeAll(t.T(), w, first)
	writeAll(t.T(), w, second)

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)
	t.Require().Len(plain, 1)
	t.Require().Equal(first, readPart(t.T(), filepath.Join(t.dir, plain[0])))
	t.Require().Equal(second, readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Rotation_KeepsAllParts() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1, MaxBackups: 0})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	p1, p2, p3 := makePayload('a', 600), makePayload('b', 600), makePayload('c', 600)
	writeAll(t.T(), w, p1)
	writeAll(t.T(), w, p2)
	writeAll(t.T(), w, p3)

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)
	t.Require().Len(plain, 2)

	got := partContents(t.T(), t.dir, plain)
	sort.Strings(got)
	want := []string{string(p1), string(p2)}
	sort.Strings(want)
	t.Require().Equal(want, got)
	t.Require().Equal(p3, readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Rotation_BackupNameFormat() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	writeAll(t.T(), w, makePayload('a', 3*1024))
	writeAll(t.T(), w, makePayload('b', 10))

	plain, _ := listBackups(t.T(), t.dir, t.name)
	t.Require().Len(plain, 1)
	t.Require().Regexp(backupNamePattern, plain[0])
}

func (t *rotatingWriterSuite) Test_Write_AppendToExistingPart() {
	pre := makePayload('x', 400)
	t.Require().NoError(os.WriteFile(t.name, pre, 0600))

	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	payload := makePayload('y', 500)
	writeAll(t.T(), w, payload)

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(plain)
	t.Require().Empty(gz)
	t.Require().Equal(append(append([]byte{}, pre...), payload...), readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Write_RotateExistingPartOverLimit() {
	pre := makePayload('x', 900)
	t.Require().NoError(os.WriteFile(t.name, pre, 0600))

	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	payload := makePayload('y', 200)
	writeAll(t.T(), w, payload)

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)
	t.Require().Len(plain, 1)
	t.Require().Equal(pre, readPart(t.T(), filepath.Join(t.dir, plain[0])))
	t.Require().Equal(payload, readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Write_OversizedNotSplit() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	big := makePayload('a', 3000)
	writeAll(t.T(), w, big)
	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(plain)
	t.Require().Empty(gz)
	t.Require().Equal(big, readPart(t.T(), t.name))

	small := makePayload('b', 10)
	writeAll(t.T(), w, small)

	plain, gz = listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)
	t.Require().Len(plain, 1)
	t.Require().Equal(big, readPart(t.T(), filepath.Join(t.dir, plain[0])))
	t.Require().Equal(small, readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Write_InvalidPathFail() {
	blocker := filepath.Join(t.dir, "not-a-dir")
	t.Require().NoError(os.WriteFile(blocker, makePayload('x', 1), 0600))

	w := file.NewFileWriter(file.Output{
		File:       filepath.Join(blocker, partFile),
		PartSizeKb: 1,
	})

	n, err := w.Write(makePayload('a', 10))
	t.Require().Error(err)
	t.Require().Zero(n)
}

func (t *rotatingWriterSuite) Test_MaxBackups_KeepsNewest() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1, MaxBackups: 2})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	const writeSize = 2 * 1024
	for _, b := range []byte{'a', 'b', 'c', 'd', 'e'} {
		writeAll(t.T(), w, makePayload(b, writeSize))
	}

	want := []string{string(makePayload('c', writeSize)), string(makePayload('d', writeSize))}
	sort.Strings(want)
	t.Require().Eventually(func() bool {
		plain, gz, ok := listBackupsSafe(t.dir, t.name)
		if !ok || len(gz) != 0 {
			return false
		}
		contents, ok := partContentsSafe(t.dir, plain)
		if !ok {
			return false
		}
		sort.Strings(contents)
		return slices.Equal(contents, want)
	}, 5*time.Second, 10*time.Millisecond, "backups to be pruned to the 2 newest")

	t.Require().Equal(makePayload('e', writeSize), readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Compress_RotatedParts() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1, Compress: true})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	const writeSize = 2 * 1024
	for _, b := range []byte{'a', 'b', 'c'} {
		writeAll(t.T(), w, makePayload(b, writeSize))
	}

	want := []string{string(makePayload('a', writeSize)), string(makePayload('b', writeSize))}
	sort.Strings(want)
	t.Require().Eventually(func() bool {
		plain, gz, ok := listBackupsSafe(t.dir, t.name)
		if !ok || len(plain) != 0 || len(gz) != 2 {
			return false
		}
		contents, ok := partContentsSafe(t.dir, gz)
		if !ok {
			return false
		}
		sort.Strings(contents)
		return slices.Equal(contents, want)
	}, 5*time.Second, 10*time.Millisecond, "rotated parts to be compressed")

	t.Require().Equal(makePayload('c', writeSize), readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Compress_WithMaxBackupsKeepsNewest() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1, MaxBackups: 1, Compress: true})
	defer func() {
		t.Require().NoError(w.Close())
	}()

	const writeSize = 2 * 1024
	for _, b := range []byte{'a', 'b', 'c', 'd'} {
		writeAll(t.T(), w, makePayload(b, writeSize))
	}

	want := []string{string(makePayload('c', writeSize))}
	t.Require().Eventually(func() bool {
		plain, gz, ok := listBackupsSafe(t.dir, t.name)
		if !ok || len(plain) != 0 || len(gz) != 1 {
			return false
		}
		contents, ok := partContentsSafe(t.dir, gz)
		if !ok {
			return false
		}
		return slices.Equal(contents, want)
	}, 5*time.Second, 10*time.Millisecond, "backups to be pruned and compressed")

	t.Require().Equal(makePayload('d', writeSize), readPart(t.T(), t.name))
}

func (t *rotatingWriterSuite) Test_Close_UnusedWriter() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})
	t.Require().NoError(w.Close())
	t.Require().NoError(w.Close())
}

func (t *rotatingWriterSuite) Test_Close_ThenWrite() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})

	first := makePayload('a', 100)
	writeAll(t.T(), w, first)
	t.Require().NoError(w.Close())
	t.Require().NoError(w.Close())

	second := makePayload('b', 100)
	writeAll(t.T(), w, second)
	t.Require().NoError(w.Close())

	t.Require().Equal(append(append([]byte{}, first...), second...), readPart(t.T(), t.name))

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(plain)
	t.Require().Empty(gz)
}

func (t *rotatingWriterSuite) Test_Close_MillGoroutineStopsAndRestarts() {
	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})

	writeAll(t.T(), w, makePayload('a', 3*1024))
	writeAll(t.T(), w, makePayload('b', 3*1024))
	t.Require().NoError(w.Close())

	t.Require().Eventually(func() bool {
		return !millGoroutineRunning()
	}, 2*time.Second, 10*time.Millisecond, "mill goroutine to stop after Close")

	writeAll(t.T(), w, makePayload('c', 3*1024))
	writeAll(t.T(), w, makePayload('d', 3*1024))
	t.Require().Eventually(millGoroutineRunning,
		2*time.Second, 10*time.Millisecond, "mill goroutine to restart on rotation after Close")
	t.Require().NoError(w.Close())
}

func (t *rotatingWriterSuite) Test_ConcurrentWrites_NoDataLost() {
	const (
		goroutines   = 10
		perGoroutine = 20
		writeSize    = 1024
	)

	w := file.NewFileWriter(file.Output{File: t.name, PartSizeKb: 1})

	payloads := make([][]byte, goroutines)
	for g := range payloads {
		payloads[g] = makePayload('A'+byte(g), writeSize)
	}

	var wg sync.WaitGroup
	for g := range payloads {
		wg.Go(func() {
			for range perGoroutine {
				_, err := w.Write(payloads[g])
				if err != nil {
					t.T().Errorf("write payload of goroutine %d: %s", g, err)
					return
				}
			}
		})
	}
	wg.Wait()
	t.Require().NoError(w.Close())

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)

	counts := make(map[byte]int, goroutines)
	for _, part := range append(plain, filepath.Base(t.name)) {
		for _, b := range readPart(t.T(), filepath.Join(t.dir, part)) {
			counts[b]++
		}
	}

	for g := range payloads {
		t.Require().Equalf(perGoroutine*writeSize, counts['A'+byte(g)], "payload of goroutine %d", g)
	}
}

type fileOutputSuite struct {
	suite.Suite

	test *test.Test

	dir  string
	name string
}

func (t *fileOutputSuite) SetupTest() {
	t.test, _ = test.New(t.T())
	t.dir = t.test.T().TempDir()
	t.name = filepath.Join(t.dir, partFile)
}

func TestFileOutput(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(fileOutputSuite))
}

func (t *fileOutputSuite) Test_ConfigToUrl_HappyPath() {
	u := file.ConfigToUrl(file.Output{
		File:       "/var/log/app.log",
		MaxSizeMb:  10,
		PartSizeKb: 218,
		MaxDays:    3,
		MaxBackups: 4,
		Compress:   true,
	})
	t.Require().Equal("lumberjack", u.Scheme)

	values, err := url.ParseQuery(u.RawQuery)
	t.Require().NoError(err)
	t.Require().Equal(map[string][]string{
		"file":       {"/var/log/app.log"},
		"maxSizeMb":  {"10"},
		"partSizeKb": {"218"},
		"maxDays":    {"3"},
		"maxBackups": {"4"},
		"compress":   {"true"},
	}, map[string][]string(values))
}

func (t *fileOutputSuite) Test_LegacyUrl_HappyPath() {
	cases := []struct {
		name       string
		partSizeKb string
	}{
		{name: "missing", partSizeKb: ""},
		{name: "notANumber", partSizeKb: "abc"},
		{name: "zero", partSizeKb: "0"},
		{name: "negative", partSizeKb: "-4"},
	}
	for _, tc := range cases {
		query := url.Values{
			"file":       {t.name},
			"maxSizeMb":  {"512"},
			"maxDays":    {"0"},
			"maxBackups": {"4"},
			"compress":   {"true"},
		}
		if tc.partSizeKb != "" {
			query.Set("partSizeKb", tc.partSizeKb)
		}
		u := &url.URL{Scheme: "lumberjack", RawQuery: query.Encode()}

		logger, err := log.NewFromConfig(log.Config{
			OutputPaths:          []string{u.String()},
			DisableDefaultOutput: true,
			InitialLevel:         log.InfoLevel,
		})
		t.Require().NoErrorf(err, "case %s", tc.name)

		logger.Info(t.T().Context(), "legacy line")
		t.Require().NoErrorf(logger.Sync(), "case %s", tc.name)

		t.Require().Containsf(string(readPart(t.T(), t.name)), "legacy line", "case %s", tc.name)

		plain, gz := listBackups(t.T(), t.dir, t.name)
		t.Require().Emptyf(plain, "case %s", tc.name)
		t.Require().Emptyf(gz, "case %s", tc.name)
	}
}

func (t *fileOutputSuite) Test_EndToEnd_ThroughKitLogger() {
	logger, err := log.New(
		log.WithDisableDefaultOutput(),
		log.WithFileOutput(file.Output{File: t.name, PartSizeKb: 1, MaxBackups: 0}),
	)
	t.Require().NoError(err)

	const messages = 100
	for i := range messages {
		logger.Info(t.T().Context(), fmt.Sprintf("message-%02d", i))
	}
	t.Require().NoError(logger.Sync())

	plain, gz := listBackups(t.T(), t.dir, t.name)
	t.Require().Empty(gz)
	t.Require().NotEmpty(plain)

	parts := append(slices.Clone(plain), filepath.Base(t.name))
	content := strings.Join(partContents(t.T(), t.dir, parts), "")

	for i := range messages {
		msg := fmt.Sprintf(`"message-%02d"`, i)
		t.Require().Equalf(1, strings.Count(content, msg), "message %d must appear exactly once", i)
	}
}

func writeAll(tb testing.TB, w io.Writer, payload []byte) {
	tb.Helper()
	n, err := w.Write(payload)
	require.NoError(tb, err)
	require.Len(tb, payload, n)
}

func makePayload(b byte, size int) []byte {
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = b
	}
	return payload
}

func readPart(tb testing.TB, path string) []byte {
	tb.Helper()
	data, err := os.ReadFile(path)
	require.NoError(tb, err)
	return data
}

func readPartSafe(path string) ([]byte, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return data, true
}

func readGzPartSafe(path string) ([]byte, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer func() {
		_ = f.Close()
	}()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, false
	}

	data, err := io.ReadAll(gz)
	if err != nil {
		return nil, false
	}
	return data, true
}

func partPrefix(current string) string {
	base := filepath.Base(current)
	return strings.TrimSuffix(base, filepath.Ext(base)) + "-"
}

// millGoroutineRunning reports whether the rotating writer cleanup goroutine
// is alive. It matches on the dot-free function name suffix, because the Go
// runtime URL-encodes dots in package paths in stack traces.
func millGoroutineRunning() bool {
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	return strings.Contains(string(buf), "rotatingFileWriter).millRun")
}

func listBackups(tb testing.TB, dir, current string) (plain, gz []string) {
	tb.Helper()
	plain, gz, ok := listBackupsSafe(dir, current)
	require.True(tb, ok)
	return plain, gz
}

// listBackupsSafe lists backup parts in dir, excluding the current part.
// It reports ok=false when the directory is unavailable so the caller can
// retry instead of failing.
func listBackupsSafe(dir, current string) (plain, gz []string, ok bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, false
	}

	prefix := partPrefix(current)
	plain = make([]string, 0, len(entries))
	gz = make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".gz") {
			gz = append(gz, entry.Name())
			continue
		}
		plain = append(plain, entry.Name())
	}
	sort.Strings(plain)
	sort.Strings(gz)
	return plain, gz, true
}

func partContents(tb testing.TB, dir string, names []string) []string {
	tb.Helper()
	contents, ok := partContentsSafe(dir, names)
	require.True(tb, ok)
	return contents
}

// partContentsSafe reads the content of the given parts, transparently
// decompressing .gz ones. It reports ok=false when a part disappears or is
// still being written, so the caller can retry instead of failing.
func partContentsSafe(dir string, names []string) (contents []string, ok bool) {
	contents = make([]string, len(names))
	for i, name := range names {
		path := filepath.Join(dir, name)
		var data []byte
		if strings.HasSuffix(name, ".gz") {
			data, ok = readGzPartSafe(path)
		} else {
			data, ok = readPartSafe(path)
		}
		if !ok {
			return nil, false
		}
		contents[i] = string(data)
	}
	return contents, true
}
