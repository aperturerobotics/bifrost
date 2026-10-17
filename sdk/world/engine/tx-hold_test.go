//go:build !js

package sdk_world_engine_test

import (
	"context"
	"strings"
	"testing"

	sdk_world_engine "github.com/s4wave/spacewave/sdk/world/engine"
	"github.com/sirupsen/logrus"
	logrus_test "github.com/sirupsen/logrus/hooks/test"
)

// TestSlowWriteNamesOpener checks that a write transaction held past the
// threshold logs the code that opened it, and a read transaction logs nothing.
func TestSlowWriteNamesOpener(t *testing.T) {
	// Log every write and capture the standard logger's entries.
	ctx := context.Background()
	engine, cleanup := setupSDKEngine(ctx, t)
	defer cleanup()
	defer sdk_world_engine.SetSlowWriteHold(0)()
	hook := logrus_test.NewGlobal()
	defer logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	// Open and discard a read, then open and commit a write.
	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	read.Discard()
	write, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := write.Commit(ctx); err != nil {
		t.Fatal(err.Error())
	}
	write.Discard()

	// Expect one entry naming this test as the opener.
	var slow []*logrus.Entry
	for _, entry := range hook.AllEntries() {
		if entry.Message == "slow World write transaction" {
			slow = append(slow, entry)
		}
	}
	if len(slow) != 1 {
		t.Fatalf("got %d slow write entries, want 1", len(slow))
	}
	if outcome := slow[0].Data["outcome"]; outcome != "commit" {
		t.Fatalf("outcome = %v, want commit", outcome)
	}
	opener, _ := slow[0].Data["opener"].(string)
	if !strings.HasPrefix(opener, "github.com/s4wave/spacewave/sdk/world/engine_test.TestSlowWriteNamesOpener:") {
		t.Fatalf("opener = %q, want this test first", opener)
	}
}
