package kernel

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"unicode/utf8"
)

// stdinReader is the console's input. A background goroutine reads the input
// source into a buffer, so poll can tell whether a read would block. It is
// also the terminal's input layer: with ISIG set it removes the signal
// characters and signals the foreground processes at once, and on a virtual
// terminal it applies the line discipline.
type stdinReader struct {
	kernel *Kernel
	// cancel stops the background goroutine.
	cancel context.CancelFunc
	// line holds the canonical line a virtual terminal is editing. Only the
	// background goroutine uses it.
	line []byte

	// mu guards the fields below.
	mu sync.Mutex
	// buf holds the input no read has taken yet.
	buf bytes.Buffer
	// eofMarks counts the end-of-file marks typed on a virtual terminal.
	// Each makes one read return end of file without ending the input.
	eofMarks int
	// eof is set when the input source has ended.
	eof bool
	// readyCh closes when a read would return at once. A read that drains
	// the buffer replaces a closed channel.
	readyCh chan struct{}
}

// newStdinReader creates the console input of k. It reads nothing until
// start.
func newStdinReader(k *Kernel) *stdinReader {
	return &stdinReader{kernel: k, cancel: func() {}, readyCh: make(chan struct{})}
}

// start reads r in the background until ctx ends, close is called or r
// fails.
func (sr *stdinReader) start(ctx context.Context, r io.Reader) {
	ctx, sr.cancel = context.WithCancel(ctx)
	go sr.readLoop(ctx, r)
}

// close stops the background read.
func (sr *stdinReader) close() {
	sr.cancel()
}

// readLoop moves input from r through the terminal's input processing into
// the buffer. An error from r ends the input.
func (sr *stdinReader) readLoop(ctx context.Context, r io.Reader) {
	dat := make([]byte, 256)
	for ctx.Err() == nil {
		n, err := r.Read(dat)
		in := sr.takeSignalChars(dat[:n])
		eofMarks := 0
		if sr.kernel.virtTTY {
			in, eofMarks = sr.discipline(sr.kernel, in)
		}
		sr.push(in, eofMarks, err != nil)
		if err != nil {
			return
		}
	}
}

// push adds processed input to the buffer and wakes the waiting readers.
func (sr *stdinReader) push(in []byte, eofMarks int, eof bool) {
	// Buffer the input.
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.buf.Write(in)
	sr.eofMarks += eofMarks
	sr.eof = sr.eof || eof

	// Wake the readers once a read would return.
	if !sr.readyLocked() {
		return
	}
	select {
	case <-sr.readyCh:
	default:
		close(sr.readyCh)
	}
}

// read takes up to len(p) bytes of input, waiting until some arrives or ctx
// ends. It returns io.EOF at an end-of-file mark or once the input ended.
func (sr *stdinReader) read(ctx context.Context, p []byte) (int, error) {
	for {
		// Take buffered input or an end of file.
		n, eof, readyCh := sr.take(p)
		if n > 0 {
			return n, nil
		}
		if eof {
			return 0, io.EOF
		}

		// Wait for more.
		if !wait(ctx, readyCh) {
			return 0, ctx.Err()
		}
	}
}

// take moves buffered input to p, or consumes an end-of-file mark. It returns
// the bytes taken, whether the read is at end of file, and the channel that
// closes when a read would next return.
func (sr *stdinReader) take(p []byte) (int, bool, <-chan struct{}) {
	// Take the input.
	sr.mu.Lock()
	defer sr.mu.Unlock()
	n, _ := sr.buf.Read(p)
	eof := sr.eof
	if n == 0 && sr.eofMarks > 0 {
		sr.eofMarks--
		eof = true
	}

	// Replace a closed ready channel once nothing is left to read.
	if !sr.readyLocked() {
		select {
		case <-sr.readyCh:
			sr.readyCh = make(chan struct{})
		default:
		}
	}
	return n, eof, sr.readyCh
}

// ready reports whether a read would return at once, and returns the channel
// that closes when one would.
func (sr *stdinReader) ready() (bool, <-chan struct{}) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.readyLocked(), sr.readyCh
}

// readyLocked reports whether a read would return at once. The caller holds
// mu.
func (sr *stdinReader) readyLocked() bool {
	return sr.buf.Len() > 0 || sr.eofMarks > 0 || sr.eof
}

// takeSignalChars removes the signal characters from in while ISIG is set,
// and acts on each at once, as the Linux terminal does even while no process
// reads the terminal. It reuses in's storage.
func (sr *stdinReader) takeSignalChars(in []byte) []byte {
	// Without ISIG, every byte is input.
	k := sr.kernel
	tio := k.termios()
	if tio.lflag()&isig == 0 {
		return in
	}

	// Keep the other bytes in order.
	cc := tio.cc()
	kept := in[:0]
	for _, b := range in {
		if sig := signalForChar(b, cc); sig != 0 {
			sr.signalChar(sig, tio.lflag()&echo != 0)
			continue
		}
		kept = append(kept, b)
	}
	return kept
}

// signalChar acts on a typed signal character: it echoes it, discards the
// line a virtual terminal is editing, and signals the foreground processes.
func (sr *stdinReader) signalChar(sig uint32, echoing bool) {
	// Echo the character.
	k := sr.kernel
	if echoing {
		_, _ = io.WriteString(k.stdout, signalCharEcho[sig])
	}

	// Discard the line, as the Linux terminal flushes its input queue.
	if k.virtTTY {
		sr.line = sr.line[:0]
	}
	k.deliverSignalToForeground(sig)
}

// signalForChar returns the signal a terminal control character sends, or
// zero for any other byte.
func signalForChar(b byte, cc []byte) uint32 {
	switch b {
	case cc[vintr]:
		return SIGINT
	case cc[vquit]:
		return SIGQUIT
	case cc[vsusp]:
		return SIGTSTP
	}
	return 0
}

// signalCharEcho is the echo of each signal's control character.
var signalCharEcho = map[uint32]string{
	SIGINT:  "^C\n",
	SIGQUIT: "^\\\n",
	SIGTSTP: "^Z\n",
}

// discipline applies a virtual terminal's line discipline to input whose
// signal characters are gone. It returns the bytes readers receive now and
// the number of end-of-file marks typed.
//
// In canonical mode, input collects in the edited line, which readers
// receive when a newline or VEOF ends it. VERASE removes the last character
// and VKILL the whole line. VEOF on an empty line is an end-of-file mark.
// In raw mode, input passes through. Input is echoed when ECHO is set.
func (sr *stdinReader) discipline(k *Kernel, data []byte) ([]byte, int) {
	// Apply the discipline to each byte.
	tio := k.termios()
	var in disciplined
	for _, b := range data {
		if b == '\r' && tio.iflag()&icrnl != 0 {
			b = '\n'
		}
		if tio.lflag()&icanon == 0 {
			in.out = append(in.out, b)
			in.echo = append(in.echo, b)
			continue
		}
		sr.editLine(&in, &tio, b)
	}

	// Echo the input.
	if tio.lflag()&echo != 0 && len(in.echo) != 0 {
		_, _ = k.stdout.Write(in.echo)
	}
	return in.out, in.eofMarks
}

// disciplined is the result of the line discipline on a run of input.
type disciplined struct {
	// out holds the bytes readers receive.
	out []byte
	// echo holds the terminal's echo when ECHO is set.
	echo []byte
	// eofMarks counts the end-of-file marks.
	eofMarks int
}

// editLine applies one byte of canonical input to the edited line. ECHOE
// echoes an erase as backspace, space, backspace.
func (sr *stdinReader) editLine(in *disciplined, tio *termios, b byte) {
	cc := tio.cc()
	eraseEcho := ""
	if tio.lflag()&echoe != 0 {
		eraseEcho = "\b \b"
	}
	switch {
	case b == cc[verase] || b == '\b':
		_, size := utf8.DecodeLastRune(sr.line)
		if size != 0 {
			sr.line = sr.line[:len(sr.line)-size]
			in.echo = append(in.echo, eraseEcho...)
		}
	case b == cc[vkill]:
		in.echo = append(in.echo, strings.Repeat(eraseEcho, utf8.RuneCount(sr.line))...)
		sr.line = sr.line[:0]
	case b == cc[veof] && len(sr.line) == 0:
		in.eofMarks++
	case b == cc[veof]:
		in.out = append(in.out, sr.line...)
		sr.line = sr.line[:0]
	case b == '\n':
		in.out = append(append(in.out, sr.line...), '\n')
		in.echo = append(in.echo, '\n')
		sr.line = sr.line[:0]
	default:
		sr.line = append(sr.line, b)
		in.echo = append(in.echo, b)
	}
}
