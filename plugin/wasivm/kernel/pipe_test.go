package kernel

import (
	"context"
	"testing"
)

// TestPipeWriteCountsAllBytes checks a write to a pipe with an open read end
// takes every byte.
func TestPipeWriteCountsAllBytes(t *testing.T) {
	// Open a pipe.
	pipe := newBufferedPipe(1)

	// Write to it with no reader waiting.
	n, ok := pipe.write(context.Background(), []byte("hello"), nil)
	if !ok {
		t.Fatal("write failed with EPIPE while the read end is open")
	}
	if n != len("hello") {
		t.Fatalf("write took %d bytes, want %d", n, len("hello"))
	}
}

// TestPipeDrainsBeforeEOF checks a pipe whose writers closed returns its
// buffered bytes before end of file.
func TestPipeDrainsBeforeEOF(t *testing.T) {
	// Buffer five bytes, then close the only writer.
	ctx := context.Background()
	pipe := newBufferedPipe(1)
	if n, ok := pipe.write(ctx, []byte("hello"), nil); !ok || n != len("hello") {
		t.Fatalf("write = %d, %t; want %d, true", n, ok, len("hello"))
	}
	pipe.close(false)

	// Read them three bytes at a time, then expect end of file.
	buf := make([]byte, 3)
	for _, want := range []string{"hel", "lo"} {
		n, _, eof := pipe.read(ctx, buf)
		if eof || string(buf[:n]) != want {
			t.Fatalf("read = %q, eof %t; want %q", buf[:n], eof, want)
		}
	}
	if n, _, eof := pipe.read(ctx, buf); !eof || n != 0 {
		t.Fatalf("read of a drained pipe = %d, eof %t; want 0, eof", n, eof)
	}
}

// TestPipeWriteAfterReadCloseFails checks a write to a pipe with no read end
// fails with EPIPE.
func TestPipeWriteAfterReadCloseFails(t *testing.T) {
	// Close the read end of a pipe and give a process the write end.
	pipe := newBufferedPipe(1)
	pipe.close(true)
	p := &Process{fds: map[uint32]*OpenFile{3: {tx: pipe, refs: 1}}}

	// Write through the descriptor.
	if got := p.writeOutBytes(context.Background(), 3, []byte("hello")); got != negErrno(EPIPE) {
		t.Fatalf("write = %#x, want EPIPE", got)
	}
	if p.sigTerm != SIGPIPE {
		t.Fatalf("terminating signal = %d, want SIGPIPE", p.sigTerm)
	}
}

// TestPipeWriteWaitsForRoom checks a write to a full pipe takes nothing until
// a read makes room, so a writer whose reader stops reading blocks.
func TestPipeWriteWaitsForRoom(t *testing.T) {
	// Fill a pipe.
	ctx := context.Background()
	pipe := newBufferedPipe(1)
	if n, ok := pipe.write(ctx, make([]byte, pipeCapacity), nil); !ok || n != pipeCapacity {
		t.Fatalf("write to fill = %d, %t; want %d, true", n, ok, pipeCapacity)
	}

	// A write that gives up takes nothing.
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if n, ok := pipe.write(canceled, []byte("x"), nil); !ok || n != 0 {
		t.Fatalf("write to a full pipe = %d, %t; want 0, true", n, ok)
	}

	// After a read, the write fits.
	if n, _, eof := pipe.read(ctx, make([]byte, 1)); eof || n != 1 {
		t.Fatalf("read = %d, eof %t; want 1, false", n, eof)
	}
	if n, ok := pipe.write(ctx, []byte("x"), nil); !ok || n != 1 {
		t.Fatalf("write after a read = %d, %t; want 1, true", n, ok)
	}
}
