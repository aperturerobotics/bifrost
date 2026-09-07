//go:build js && !tinygo

package opfs

import (
	"io"
	"syscall/js"
	"testing"
)

// TestReadSnapshotWholeAndPartialRanges checks immutable File reads and EOF
// while ensuring complete reads do not allocate an intermediate Blob.
func TestReadSnapshotWholeAndPartialRanges(t *testing.T) {
	// Retain a real browser File and count only its slice operations.
	const content = "immutable snapshot bytes"
	file := js.Global().Get("File").New([]any{content}, "snapshot.bin")
	originalSlice := file.Get("slice")
	sliceCalls := 0
	slice := js.FuncOf(func(this js.Value, args []js.Value) any {
		sliceCalls++
		return originalSlice.Call("call", this, args[0], args[1])
	})
	t.Cleanup(slice.Release)
	file.Set("slice", slice)
	snapshot := &ReadSnapshot{
		driver: BrowserDriver{},
		name:   "snapshot.bin",
		handle: file,
		size:   int64(len(content)),
	}
	t.Cleanup(func() {
		if err := snapshot.Close(); err != nil {
			t.Error(err)
		}
	})

	// Both exact and oversized reads use the complete immutable File.
	for _, extra := range []int{0, 5} {
		buf := make([]byte, len(content)+extra)
		n, err := snapshot.ReadAt(buf, 0)
		if n != len(content) || string(buf[:n]) != content {
			t.Fatalf("whole read = (%d, %q), want %q", n, buf[:n], content)
		}
		wantErr := error(nil)
		if extra != 0 {
			wantErr = io.EOF
		}
		if err != wantErr {
			t.Fatalf("whole read error = %v, want %v", err, wantErr)
		}
	}
	if sliceCalls != 0 {
		t.Fatalf("whole reads made %d slice calls", sliceCalls)
	}

	// Partial reads retain the range operation and return only requested bytes.
	buf := make([]byte, 4)
	n, err := snapshot.ReadAt(buf, 2)
	if n != len(buf) || err != nil || string(buf) != content[2:6] {
		t.Fatalf("partial read = (%d, %q, %v)", n, buf, err)
	}
	if sliceCalls != 1 {
		t.Fatalf("partial read made %d slice calls, want 1", sliceCalls)
	}
}
