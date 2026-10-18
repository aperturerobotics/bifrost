//go:build !js

package spacewave_cli

import "testing"

// TestSpaceBackfillHelpDescription checks the exact backfill help description.
func TestSpaceBackfillHelpDescription(t *testing.T) {
	if got, want := newSpaceBackfillCommand(nil, nil).Description, "With backfill on, this device copies the installed World and its retained roots into its local store in the background. With backfill off, the default, it fetches blocks on demand into an evictable cache. Storage reclaim can remove cached blocks unless a retained root reaches them. Without on or off, it prints the current choice."; got != want {
		t.Fatalf("backfill help description = %q, want %q", got, want)
	}
}
