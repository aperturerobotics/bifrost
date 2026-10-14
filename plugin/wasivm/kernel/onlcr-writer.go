package kernel

import (
	"io"
	"sync"
)

// onlcrWriter writes the console's output to a host stream after the
// console's output processing.
type onlcrWriter struct {
	// term is the console.
	term *tty
	// inner is the host stream.
	inner io.Writer
	// mu serializes Write. One writer serves every process on the console,
	// each on its own goroutine, so the lock makes each guest write atomic
	// on the host stream, as Linux write is.
	mu sync.Mutex
}

// Write processes p and writes it to the host stream, reporting the bytes of
// p written.
func (w *onlcrWriter) Write(p []byte) (int, error) {
	// Write the processed bytes and report the caller's bytes written.
	w.mu.Lock()
	defer w.mu.Unlock()
	out := w.term.processOutput(p)
	n, err := w.inner.Write(out)
	return inputBytes(len(p), n, len(out)), err
}

// _ is a type assertion
var _ io.Writer = (*onlcrWriter)(nil)
