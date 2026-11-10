package file

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/txix-open/isp-kit/errors"
)

const (
	// backupTimeFormat is the timestamp format for rotated log file names.
	// It matches the lumberjack format so that backup files produced by both
	// writers are interchangeable in the same directory.
	backupTimeFormat = "2006-01-02T15-04-05.000"

	compressSuffix = ".gz"

	kilobytes   = int64(1) << 10
	logDirMode  = os.FileMode(0o755)
	logFileMode = os.FileMode(0o600)
)

// rotatingFileWriter is an io.WriteCloser that writes to the given log file
// and rotates it whenever a write would make the file larger than the
// configured part size in bytes.
//
// It complements lumberjack, which only supports whole-megabyte size limits,
// and is used for byte-precise part sizes (e.g. 218 KB for VictoriaLogs).
//
// A single write larger than the part size is not rejected: a rotation
// happens first and then the write is made to the new part, so the part may
// exceed the limit by at most one log line.
//
// The writer assumes that only one process writes to the log file.
type rotatingFileWriter struct {
	filename   string
	maxSizeB   int64
	maxBackups int
	compress   bool

	mu   sync.Mutex
	file *os.File
	size int64

	millCh chan struct{} // cleanup signals; nil while the mill is stopped
	done   chan struct{} // closed when the mill goroutine exits
}

// newRotatingFileWriter creates a new rotatingFileWriter.
//
// Parameters:
//   - filename: path to the log file; backup parts are kept in the same directory
//   - maxSizeB: maximum part size in bytes
//   - maxBackups: maximum number of backup parts to keep (0 means keep all)
//   - compress: whether rotated parts are compressed with gzip
func newRotatingFileWriter(filename string, maxSizeB int64, maxBackups int, compress bool) *rotatingFileWriter {
	return &rotatingFileWriter{
		filename:   filename,
		maxSizeB:   maxSizeB,
		maxBackups: maxBackups,
		compress:   compress,
	}
}

// Write implements io.Writer. If a write would make the log file larger than
// the part size, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created with the original name.
func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		err := w.openExistingOrNew(int64(len(p)))
		if err != nil {
			return 0, err
		}
	}

	// A rotation happens only for a non-empty part, so an oversized first
	// write does not produce an empty backup part.
	if w.size > 0 && w.size+int64(len(p)) > w.maxSizeB {
		err := w.rotate()
		if err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	if err == nil {
		w.size += int64(n)
	}
	return n, err
}

// Close implements io.Closer and closes the current log file, stopping the
// post-rotation cleanup. The next Write opens the file again, and the cleanup
// starts again on the next rotation.
func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	err := w.close()
	w.stopMill()
	return err
}

// close closes the file if it is open.
// The caller must hold the lock.
func (w *rotatingFileWriter) close() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

// openExistingOrNew opens the log file if it exists and a write of the given
// length would not put it over the part size. Otherwise a rotation happens
// and a new part is created.
// The caller must hold the lock.
func (w *rotatingFileWriter) openExistingOrNew(writeLen int64) error {
	info, err := os.Stat(w.filename)
	if os.IsNotExist(err) {
		return w.openNew()
	}
	if err != nil {
		return errors.WithMessage(err, "error getting log file info")
	}

	if info.Size() > 0 && info.Size()+writeLen > w.maxSizeB {
		return w.rotate()
	}

	file, err := os.OpenFile(w.filename, os.O_APPEND|os.O_WRONLY, logFileMode)
	if err != nil {
		// If we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return w.openNew()
	}
	w.file = file
	w.size = info.Size()
	return nil
}

// rotate closes the current file, moves it aside under a timestamped backup
// name, opens a new part with the original name and starts the post-rotation
// cleanup.
// The caller must hold the lock.
func (w *rotatingFileWriter) rotate() error {
	err := w.close()
	if err != nil {
		return err
	}
	err = w.openNew()
	if err != nil {
		return err
	}
	w.mill()
	return nil
}

// openNew opens a new log part for writing, moving any old log file out of
// the way under a timestamped backup name. The file must be closed before
// calling.
// The caller must hold the lock.
func (w *rotatingFileWriter) openNew() error {
	err := os.MkdirAll(filepath.Dir(w.filename), logDirMode)
	if err != nil {
		return errors.WithMessage(err, "can't make directories for new logfile")
	}

	mode := logFileMode
	info, err := os.Stat(w.filename)
	switch {
	case err == nil:
		// Copy the mode off the old logfile.
		mode = info.Mode()
		err = os.Rename(w.filename, w.backupName())
		if err != nil {
			return errors.WithMessage(err, "can't rename log file")
		}
	case !os.IsNotExist(err):
		return errors.WithMessage(err, "error getting log file info")
	}

	f, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return errors.WithMessage(err, "can't open new logfile")
	}
	w.file = f
	w.size = 0
	return nil
}

// backupName creates a new backup name from the log file name, inserting a UTC
// timestamp between the filename and the extension. If a part rotated at the
// same moment exists, a numeric collision suffix is inserted before the
// extension.
func (w *rotatingFileWriter) backupName() string {
	dir := filepath.Dir(w.filename)
	base := filepath.Base(w.filename)
	ext := filepath.Ext(base)
	prefix := base[:len(base)-len(ext)] + "-"

	now := time.Now().UTC()
	stamp := now.Format(backupTimeFormat)
	name := prefix + stamp

	collision, ok := nextCollisionNumber(dir, prefix, ext, stamp)
	if !ok {
		// The directory cannot be listed, so fall back to probing for a free name.
		target := filepath.Join(dir, name+ext)
		for i := 1; fileExists(target); i++ {
			target = filepath.Join(dir, fmt.Sprintf("%s-%d%s", name, i, ext))
		}
		return target
	}
	if collision == 0 {
		return filepath.Join(dir, name+ext)
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%d%s", name, collision, ext))
}

// nextCollisionNumber returns the next collision number for a part rotated at
// the given timestamp: the maximum existing number plus one, or zero when no
// part with the same timestamp exists.
// The number grows monotonically within the same timestamp: the mill may free
// a smaller number when it removes or compresses an older part, and reusing it
// would make a newer part look older than a surviving one, so the cleanup would
// keep the stale part and remove the fresh one.
func nextCollisionNumber(dir, prefix, ext, stamp string) (int, bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, false
	}

	collision := -1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		timestamp, number, ok := parseBackupName(entry.Name(), prefix, ext)
		if !ok || timestamp.Format(backupTimeFormat) != stamp {
			continue
		}
		if number > collision {
			collision = number
		}
	}
	return collision + 1, true
}

// mill schedules the post-rotation cleanup (compression and removal of stale
// backup parts), starting the cleanup goroutine if necessary. Frequent
// rotations coalesce: every cleanup pass scans all backup parts.
// The caller must hold the lock.
func (w *rotatingFileWriter) mill() {
	if w.millCh == nil {
		w.millCh = make(chan struct{}, 1)
		w.done = make(chan struct{})
		go w.millRun()
	}
	select {
	case w.millCh <- struct{}{}:
	default:
	}
}

// millRun runs in a goroutine to manage post-rotation compression and
// removal of old backup parts. It exits when millCh is closed, after
// processing a pending cleanup signal.
func (w *rotatingFileWriter) millRun() {
	defer close(w.done)
	for range w.millCh {
		_ = w.millRunOnce()
	}
}

// stopMill stops the post-rotation cleanup and waits for the goroutine to
// exit.
// The caller must hold the lock.
func (w *rotatingFileWriter) stopMill() {
	if w.millCh == nil {
		return
	}
	close(w.millCh)
	<-w.done
	w.millCh = nil
	w.done = nil
}

// millRunOnce removes the oldest backup parts keeping at most maxBackups
// (0 means keep all) and compresses uncompressed parts if compression is
// enabled.
func (w *rotatingFileWriter) millRunOnce() error {
	if w.maxBackups == 0 && !w.compress {
		return nil
	}

	files, err := w.oldBackups()
	if err != nil {
		return err
	}

	remove, compress := w.planCleanup(files)
	if len(remove) == 0 && len(compress) == 0 {
		return nil
	}

	var firstErr error
	for _, f := range remove {
		err := os.Remove(f)
		if err != nil && firstErr == nil {
			firstErr = errors.WithMessage(err, "can't remove old log file")
		}
	}
	for _, f := range compress {
		err := compressFile(f, f+compressSuffix)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// planCleanup splits the backup parts (newest first) into the files to
// remove and the files to compress.
func (w *rotatingFileWriter) planCleanup(files []backupFile) (remove, compress []string) {
	kept := 0
	seen := make(map[string]bool, len(files))
	for _, f := range files {
		key, _ := strings.CutSuffix(f.path, compressSuffix)
		if !seen[key] {
			seen[key] = true
			kept++
		}
		if w.maxBackups > 0 && kept > w.maxBackups {
			remove = append(remove, f.path)
		}
	}

	removed := make(map[string]bool, len(remove))
	for _, f := range remove {
		removed[f] = true
	}

	if w.compress {
		for _, f := range files {
			if !strings.HasSuffix(f.path, compressSuffix) && !removed[f.path] {
				compress = append(compress, f.path)
			}
		}
	}
	return remove, compress
}

// oldBackups returns the paths of backup log parts in the log file directory,
// sorted by the timestamp encoded in their names, newest first.
func (w *rotatingFileWriter) oldBackups() ([]backupFile, error) {
	dir := filepath.Dir(w.filename)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.WithMessage(err, "can't read log file directory")
	}

	base := filepath.Base(w.filename)
	ext := filepath.Ext(base)
	prefix := base[:len(base)-len(ext)] + "-"

	backups := []backupFile{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		timestamp, collision, ok := parseBackupName(entry.Name(), prefix, ext)
		if !ok {
			continue
		}
		backups = append(backups, backupFile{
			path:      filepath.Join(dir, entry.Name()),
			timestamp: timestamp,
			collision: collision,
		})
	}
	sort.Slice(backups, func(i, j int) bool {
		if !backups[i].timestamp.Equal(backups[j].timestamp) {
			return backups[i].timestamp.After(backups[j].timestamp)
		}
		// Parts rotated in the same millisecond carry a numeric collision
		// suffix; a larger suffix is the newer part.
		return backups[i].collision > backups[j].collision
	})

	return backups, nil
}

// backupFile is a backup log part with the timestamp and collision number
// parsed from its name.
type backupFile struct {
	path      string
	timestamp time.Time
	collision int
}

// parseBackupName extracts the rotation timestamp and collision number from a
// backup part name.
// The name must have the form "<prefix>-<timestamp><ext>" optionally followed
// by a numeric collision suffix and/or the ".gz" compression suffix.
func parseBackupName(filename, prefix, ext string) (timestamp time.Time, collision int, ok bool) {
	name, _ := strings.CutSuffix(filename, compressSuffix)
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ext) {
		return time.Time{}, 0, false
	}
	timestamp, collision, ok = parseTimestamp(name[len(prefix) : len(name)-len(ext)])
	return timestamp, collision, ok
}

// parseTimestamp parses a backup timestamp, tolerating a trailing numeric
// collision suffix ("-1", "-2", ...).
func parseTimestamp(value string) (t time.Time, collision int, ok bool) {
	if index := strings.LastIndexByte(value, '-'); index > 0 {
		number, convErr := strconv.Atoi(value[index+1:])
		parsed, parseErr := time.Parse(backupTimeFormat, value[:index])
		if convErr == nil && parseErr == nil {
			return parsed, number, true
		}
	}
	parsed, err := time.Parse(backupTimeFormat, value)
	if err != nil {
		return time.Time{}, 0, false
	}
	return parsed, 0, true
}

// compressFile compresses the given file with gzip, removing the original
// file if successful.
func compressFile(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return errors.WithMessage(err, "failed to open log file")
	}
	defer func() {
		_ = f.Close()
	}()

	fi, err := os.Stat(src)
	if err != nil {
		return errors.WithMessage(err, "failed to stat log file")
	}

	gzf, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fi.Mode())
	if err != nil {
		return errors.WithMessage(err, "failed to open compressed log file")
	}
	defer func() {
		_ = gzf.Close()
	}()

	gz := gzip.NewWriter(gzf)
	_, err = io.Copy(gz, f)
	if err != nil {
		return errors.WithMessage(err, "failed to compress log file")
	}
	err = gz.Close()
	if err != nil {
		return errors.WithMessage(err, "failed to compress log file")
	}
	err = gzf.Close()
	if err != nil {
		return errors.WithMessage(err, "failed to close compressed log file")
	}
	err = f.Close()
	if err != nil {
		return errors.WithMessage(err, "failed to close log file")
	}
	err = os.Remove(src)
	if err != nil {
		return errors.WithMessage(err, "failed to remove uncompressed log file")
	}

	return nil
}

// fileExists reports whether the file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
