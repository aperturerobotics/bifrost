package sobject_test

import (
	"bytes"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
)

// TestSharedObjectListEntryWireFields checks stored peer IDs and update timestamps.
func TestSharedObjectListEntryWireFields(t *testing.T) {
	// Decode a stored transport_peer_id in field 4 and updated_at in field 5.
	data := []byte{0x22, 0x04, 'p', 'e', 'e', 'r', 0x28, 0x96, 0x01}
	var entry sobject.SharedObjectListEntry
	if err := entry.UnmarshalVT(data); err != nil {
		t.Fatal(err)
	}

	// Read the timestamp independently of the unknown peer ID.
	if entry.GetUpdatedAt() != 150 {
		t.Fatalf("updated_at = %d, want 150", entry.GetUpdatedAt())
	}

	// Encode the timestamp with field 5's varint tag.
	encoded, err := (&sobject.SharedObjectListEntry{UpdatedAt: 150}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, []byte{0x28, 0x96, 0x01}) {
		t.Fatalf("encoded updated_at = %x, want 289601", encoded)
	}
}
