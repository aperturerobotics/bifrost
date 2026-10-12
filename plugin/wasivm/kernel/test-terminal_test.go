package kernel

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aperturerobotics/util/broadcast"
)

// testTerminal runs an interactive BusyBox shell on a virtual terminal. The
// test types input and waits for output, so each step follows the guest
// instead of a timer.
type testTerminal struct {
	// t is the running test.
	t *testing.T
	// ctx bounds the session.
	ctx context.Context
	// in feeds the terminal input.
	in *io.PipeWriter
	// done receives the shell's result.
	done chan error

	// bcast guards out and seen and wakes waiters when output arrives.
	bcast broadcast.Broadcast
	// out is the terminal output so far.
	out bytes.Buffer
	// seen is the offset in out after the last output waited for.
	seen int
}

// newTestTerminal starts an interactive shell. It skips the test when the
// BusyBox build is missing.
func newTestTerminal(t *testing.T) *testTerminal {
	// Load BusyBox.
	bbWasm := busyboxWasm(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	// Boot a kernel on a virtual terminal.
	pr, pw := io.Pipe()
	tt := &testTerminal{t: t, ctx: ctx, in: pw, done: make(chan error, 1)}
	k, err := New(ctx, testRuntime(t, ctx), &Config{
		Stdout:     tt,
		Stderr:     tt,
		Stdin:      pr,
		TerminalFD: 99,
		Files:      BusyboxApplets(bbWasm),
		Env:        map[string]string{"PATH": "/bin:/usr/bin", "HOME": "/root"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = k.Close(ctx) })

	// Run the shell.
	go func() {
		tt.done <- k.RunProcess(ctx, bbWasm, "sh")
	}()
	return tt
}

// Write records terminal output.
func (tt *testTerminal) Write(p []byte) (int, error) {
	// Append it and wake the waiters.
	l := tt.bcast.Lock()
	defer l.Unlock()
	tt.out.Write(p)
	l.Broadcast()
	return len(p), nil
}

// typeText types s on the terminal.
func (tt *testTerminal) typeText(s string) {
	if _, err := tt.in.Write([]byte(s)); err != nil {
		tt.t.Fatal(err)
	}
}

// waitFor waits until the output after the last wait contains s.
func (tt *testTerminal) waitFor(s string) {
	for {
		// Look for s in the output not yet waited for.
		l := tt.bcast.Lock()
		if i := strings.Index(tt.out.String()[tt.seen:], s); i >= 0 {
			tt.seen += i + len(s)
			l.Unlock()
			return
		}
		out := tt.out.String()
		waitCh := l.WaitCh()
		l.Unlock()

		select {
		case <-tt.ctx.Done():
			tt.t.Fatalf("waiting for %q, output: %q", s, out)
		case err := <-tt.done:
			tt.t.Fatalf("shell exited (%v) while waiting for %q, output: %q", err, s, out)
		case <-waitCh:
		}
	}
}

// exit types exit and waits for the shell to end without error.
func (tt *testTerminal) exit() {
	tt.typeText("exit\n")
	select {
	case <-tt.ctx.Done():
		tt.t.Fatal("shell did not exit")
	case err := <-tt.done:
		if err != nil {
			tt.t.Fatal(err)
		}
	}
}
