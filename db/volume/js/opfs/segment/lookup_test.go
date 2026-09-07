package segment

import (
	"bytes"
	"testing"
)

// TestLookupMetaHasStreamsLargeExistenceWindow verifies scalar existence
// lookups scan entry metadata instead of reading large values.
func TestLookupMetaHasStreamsLargeExistenceWindow(t *testing.T) {
	const (
		entryCount = 16
		valueSize  = 8 * 1024
	)

	// Build one sparse-index window above the existence buffering threshold.
	w := NewWriter()
	w.SetIndexInterval(entryCount)
	for i := range entryCount {
		w.Add(
			[]byte("key-"+zeroPad(i, 4)),
			bytes.Repeat([]byte{byte('a' + i%26)}, valueSize),
		)
	}

	var buf bytes.Buffer
	if _, err := w.Build(&buf); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Load metadata before recording data-window reads.
	data := buf.Bytes()
	meta, err := LoadLookupMeta(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("LoadLookupMeta: %v", err)
	}

	// Check bounded reads at the beginning, middle, and end of the window.
	positions := []struct {
		name string
		key  string
	}{
		{name: "first", key: "key-0000"},
		{name: "middle", key: "key-0007"},
		{name: "last", key: "key-0015"},
	}
	for _, position := range positions {
		t.Run(position.name, func(t *testing.T) {
			reader := &recordingReaderAt{data: data}
			found, err := meta.Has(reader, []byte(position.key))
			if err != nil {
				t.Fatalf("Has: %v", err)
			}
			if !found {
				t.Fatal("Has: not found")
			}
			if reader.maxRead > len(position.key) {
				t.Fatalf("largest ReadAt = %d, want at most key length %d", reader.maxRead, len(position.key))
			}
		})
	}
}
