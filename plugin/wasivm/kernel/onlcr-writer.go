package kernel

import (
	"io"
	"sync"
)

// count32 maps the bytes written of ONLCR-expanded output back to a count of
// the caller's bytes.
func count32(origLen, written, expandedLen int) int {
	if expandedLen == 0 {
		return 0
	}
	if written >= expandedLen {
		return origLen
	}
	return written * origLen / expandedLen
}

// onlcrWriter wraps an io.Writer with NL-to-CRNL output processing.
type onlcrWriter struct {
	kernel *Kernel
	inner  io.Writer
	// mu serializes Write. A single onlcrWriter backs k.stdout/k.stderr for the
	// parent and every exec child, which run on separate goroutines; the lock
	// makes each guest write() atomic on the shared inner writer, matching Linux
	// write() atomicity and preventing interleaved or racing byte slices.
	mu sync.Mutex
}

// Write applies ONLCR processing and writes to the underlying writer.
func (w *onlcrWriter) Write(p []byte) (int, error) {
	// Write the processed bytes and report the caller's bytes written.
	w.mu.Lock()
	defer w.mu.Unlock()
	out := w.kernel.onlcrProcess(p)
	n, err := w.inner.Write(out)
	if len(out) > len(p) {
		return count32(len(p), n, len(out)), err
	}
	return n, err
}

// onlcrProcess translates \n to \r\n when OPOST+ONLCR are set and the
// kernel processes output: on a virtual terminal, or on a host terminal in
// raw mode. Returns buf unchanged if no translation needed.
func (k *Kernel) onlcrProcess(buf []byte) []byte {
	// Pass the output through unless the kernel processes it with ONLCR set.
	if !k.termRaw.Load() && !k.virtTTY {
		return buf
	}
	tio := k.termios()
	if oflag := tio.oflag(); oflag&opost == 0 || oflag&onlcr == 0 {
		return buf
	}

	// Count the bare newlines.
	extra := 0
	for i, b := range buf {
		if b == '\n' && (i == 0 || buf[i-1] != '\r') {
			extra++
		}
	}
	if extra == 0 {
		return buf
	}

	// Put a carriage return before each.
	out := make([]byte, 0, len(buf)+extra)
	for i, b := range buf {
		if b == '\n' && (i == 0 || buf[i-1] != '\r') {
			out = append(out, '\r')
		}
		out = append(out, b)
	}
	return out
}
