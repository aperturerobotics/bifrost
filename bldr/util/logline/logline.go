// Package logline provides an io.Writer that logs each line a child process
// writes at the logrus level the line names.
package logline

import (
	"bytes"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Writer logs each complete line written to it through an entry. A line in
// logrus text format, starting with level=NAME, logs at that level with the
// level field removed; any other line logs at the fallback level. A trailing
// partial line waits for its newline.
type Writer struct {
	le       *logrus.Entry
	fallback logrus.Level
	mu       sync.Mutex
	buf      []byte
}

// New constructs a Writer logging through le, with lines that name no level
// logged at fallback.
func New(le *logrus.Entry, fallback logrus.Level) *Writer {
	return &Writer{le: le, fallback: fallback}
}

// Write logs every complete line in p and keeps the rest for the next write.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	for {
		idx := bytes.IndexByte(w.buf, '\n')
		if idx < 0 {
			return len(p), nil
		}
		w.logLine(string(bytes.TrimRight(w.buf[:idx], "\r")))
		w.buf = w.buf[idx+1:]
	}
}

// logLine logs one line at the level it names, or at the fallback.
func (w *Writer) logLine(line string) {
	if line == "" {
		return
	}
	level, rest := w.fallback, line
	if name, after, ok := strings.Cut(line, " "); ok && strings.HasPrefix(name, "level=") {
		if parsed, err := logrus.ParseLevel(strings.TrimPrefix(name, "level=")); err == nil {
			level, rest = parsed, after
		}
	}
	w.le.Log(level, rest)
}
