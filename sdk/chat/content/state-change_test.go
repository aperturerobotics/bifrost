package spacewave_chat_content

import (
	"strings"
	"testing"
)

// TestStateChangeIdentityBound checks that state identities are bounded by
// characters, not bytes.
func TestStateChangeIdentityBound(t *testing.T) {
	for _, tc := range []struct {
		stateKey string
		valid    bool
	}{
		{strings.Repeat("💥", 255), true},
		{strings.Repeat("x", 256), false},
		{"a\x00b", false},
		{"\xff", false},
	} {
		change := &ChatStateChange{Type: "m.room.member", StateKey: tc.stateKey, ContentJson: "{}"}
		if err := change.Validate(); (err == nil) != tc.valid {
			t.Fatalf("state key of %d bytes: got %v, want valid %v", len(tc.stateKey), err, tc.valid)
		}
	}
}
