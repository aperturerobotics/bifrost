package kernel

import (
	"context"
	"testing"
)

// TestPipeWriteCountsAllBytes checks a write to a pipe with an open read end
// takes every byte.
func TestPipeWriteCountsAllBytes(t *testing.T) {
	// Open a pipe.
	reader, writer := newPipeEnds()
	defer reader.unref()
	defer writer.unref()

	// Write to it with no reader waiting.
	n, ok := writer.write([]byte("hello"))
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
	reader, writer := newPipeEnds()
	defer reader.unref()
	if n, ok := writer.write([]byte("hello")); !ok || n != len("hello") {
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
	reader, writer := newPipeEnds()
	defer writer.unref()
	reader.unref()
	p := &Process{fds: map[uint32]*OpenFile{3: {pipe: writer}}}

	// Write through the descriptor.
	if got := p.writeOutBytes(3, []byte("hello")); got != negErrno(EPIPE) {
		t.Fatalf("write = %#x, want EPIPE", got)
	}
}

// TestPipeReadRefReopens checks a new reference to a closed read end makes
// the pipe writable again.
func TestPipeReadRefReopens(t *testing.T) {
	// Close the read end and expect writes to fail.
	reader, writer := newPipeEnds()
	defer reader.unref()
	defer writer.unref()
	reader.unref()
	if n, ok := writer.write([]byte("before")); ok || n != 0 {
		t.Fatalf("write after close = %d, %t; want 0, false", n, ok)
	}

	// Reference it again and expect writes to succeed.
	reader.ref()
	if n, ok := writer.write([]byte("after")); !ok || n != len("after") {
		t.Fatalf("write after reopen = %d, %t; want %d, true", n, ok, len("after"))
	}
}

// TestPipeHeldReadRefSurvivesClose checks a held read reference keeps the pipe
// open while the last ordinary reference closes, as across vfork.
func TestPipeHeldReadRefSurvivesClose(t *testing.T) {
	// Hold the read end, drop its reference and write.
	ctx := context.Background()
	reader, writer := newPipeEnds()
	defer reader.unref()
	defer writer.unref()
	reader.hold()
	reader.unref()
	if n, ok := writer.write([]byte("held")); !ok || n != len("held") {
		t.Fatalf("write with a held read end = %d, %t; want %d, true", n, ok, len("held"))
	}

	// Restore the held reference and read the bytes back.
	reader.restoreHeld()
	buf := make([]byte, len("held"))
	if n, eof := reader.read(ctx, buf); eof || string(buf[:n]) != "held" {
		t.Fatalf("read after restore = %q, eof %t; want %q", buf[:n], eof, "held")
	}
}
