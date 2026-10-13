package kernel

import (
	"context"
	"testing"
)

// TestPipeWriteCountsAllBytes checks a write to a pipe with an open read end
// takes every byte.
func TestPipeWriteCountsAllBytes(t *testing.T) {
	// Open a pipe.
	reader, writer := newPipeEnds(1)
	defer reader.unref()
	defer writer.unref()

	// Write to it with no reader waiting.
	n, ok := writer.write(context.Background(), []byte("hello"))
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
	reader, writer := newPipeEnds(1)
	defer reader.unref()
	if n, ok := writer.write(ctx, []byte("hello")); !ok || n != len("hello") {
		t.Fatalf("write = %d, %t; want %d, true", n, ok, len("hello"))
	}
	writer.unref()

	// Read them three bytes at a time, then expect end of file.
	buf := make([]byte, 3)
	for _, want := range []string{"hel", "lo"} {
		n, eof := reader.read(ctx, buf)
		if eof || string(buf[:n]) != want {
			t.Fatalf("read = %q, eof %t; want %q", buf[:n], eof, want)
		}
	}
	if n, eof := reader.read(ctx, buf); !eof || n != 0 {
		t.Fatalf("read of a drained pipe = %d, eof %t; want 0, eof", n, eof)
	}
}

// TestPipeWriteAfterReadCloseFails checks a write to a pipe with no read end
// fails with EPIPE.
func TestPipeWriteAfterReadCloseFails(t *testing.T) {
	// Close the read end of a pipe and give a process the write end.
	reader, writer := newPipeEnds(1)
	defer writer.unref()
	reader.unref()
	p := &Process{fds: map[uint32]*OpenFile{3: {pipe: writer}}}

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
	reader, writer := newPipeEnds(1)
	defer reader.unref()
	defer writer.unref()
	if n, ok := writer.write(ctx, make([]byte, pipeCapacity)); !ok || n != pipeCapacity {
		t.Fatalf("write to fill = %d, %t; want %d, true", n, ok, pipeCapacity)
	}

	// A write that gives up takes nothing.
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if n, ok := writer.write(canceled, []byte("x")); !ok || n != 0 {
		t.Fatalf("write to a full pipe = %d, %t; want 0, true", n, ok)
	}

	// After a read, the write fits.
	if n, eof := reader.read(ctx, make([]byte, 1)); eof || n != 1 {
		t.Fatalf("read = %d, eof %t; want 1, false", n, eof)
	}
	if n, ok := writer.write(ctx, []byte("x")); !ok || n != 1 {
		t.Fatalf("write after a read = %d, %t; want 1, true", n, ok)
	}
}
