package world

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/net/hash"
)

// readStubWorldState answers the reads ReadSet tracks with empty results.
type readStubWorldState struct {
	WorldState
}

// GetObject reports that no object exists.
func (readStubWorldState) GetObject(context.Context, string) (ObjectState, bool, error) {
	return nil, false, nil
}

// IterateObjects returns no iterator.
func (readStubWorldState) IterateObjects(context.Context, string, bool) ObjectIterator {
	return nil
}

// LookupGraphQuads returns no quads.
func (readStubWorldState) LookupGraphQuads(context.Context, GraphQuad, uint32) ([]GraphQuad, error) {
	return nil, nil
}

// BuildStorageCursor returns no cursor.
func (readStubWorldState) BuildStorageCursor(context.Context) (*bucket_lookup.Cursor, error) {
	return nil, nil
}

// AccessWorldState calls cb without a cursor.
func (readStubWorldState) AccessWorldState(_ context.Context, _ *bucket.ObjectRef, cb func(*bucket_lookup.Cursor) error) error {
	return cb(nil)
}

// TestReadSetTouched checks which ChangeSets touch the reads of a view.
func TestReadSetTouched(t *testing.T) {
	// Read key a and iterate prefix dir/.
	ctx := t.Context()
	reads := NewReadSet()
	ws := reads.Track(readStubWorldState{})
	obj, _, err := ws.GetObject(ctx, "a")
	ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	ws.IterateObjects(ctx, "dir/", false)

	// Look up (s, p, *), (*, q, o2), a literal object and (*, bare, o3).
	if _, err := ws.LookupGraphQuads(ctx, NewGraphQuadWithKeys("s", "p", "", ""), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ws.LookupGraphQuads(ctx, NewGraphQuadWithKeys("", "q", "o2", ""), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ws.LookupGraphQuads(ctx, NewGraphQuad("", "q", `"literal"`, ""), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ws.LookupGraphQuads(ctx, NewGraphQuadWithKeys("", "bare", "o3", ""), 0); err != nil {
		t.Fatal(err)
	}

	// Compare each ChangeSet against the reads.
	cases := []struct {
		name    string
		changes *ChangeSet
		touched bool
	}{
		{"empty", &ChangeSet{}, false},
		{"other key", &ChangeSet{Keys: []string{"b"}}, false},
		{"other subject", &ChangeSet{Quads: []GraphQuad{NewGraphQuad("<s2>", `"p"`, "<o>", "")}}, false},
		{"read key", &ChangeSet{Keys: []string{"a"}}, true},
		{"create under prefix", &ChangeSet{Keys: []string{"dir/x"}}, true},
		{"matching quad", &ChangeSet{Quads: []GraphQuad{NewGraphQuad("<s>", `"p"`, "<o>", "")}}, true},
		{"rename into read key", &ChangeSet{Keys: []string{"c", "a"}}, true},
		{"pattern subject", &ChangeSet{Keys: []string{"s"}}, true},
		{"pattern object", &ChangeSet{Keys: []string{"o2"}}, true},
		{"literal object", &ChangeSet{Keys: []string{"literal"}}, false},
		{"bare predicate", &ChangeSet{Quads: []GraphQuad{NewGraphQuad("<s3>", `"bare"`, "<o3>", "")}}, true},
		{"unknown", NewUnknownChangeSet(), true},
	}
	for _, c := range cases {
		if got := reads.Touched(c.changes); got != c.touched {
			t.Errorf("%s: touched = %v, want %v", c.name, got, c.touched)
		}
	}
}

// TestReadSetStorageReads checks that only a read of the root World state
// touches every change. Blocks named by a ref never change.
func TestReadSetStorageReads(t *testing.T) {
	ctx := t.Context()
	changes := &ChangeSet{Keys: []string{"other"}}
	noop := func(*bucket_lookup.Cursor) error { return nil }

	// Build a cursor and follow a ref.
	reads := NewReadSet()
	ws := reads.Track(readStubWorldState{})
	if _, err := ws.BuildStorageCursor(ctx); err != nil {
		t.Fatal(err)
	}
	h, err := hash.Sum(hash.HashType_HashType_SHA256, []byte("block"))
	if err != nil {
		t.Fatal(err)
	}
	ref := &bucket.ObjectRef{RootRef: &block.BlockRef{Hash: h}}
	if err := ws.AccessWorldState(ctx, ref, noop); err != nil {
		t.Fatal(err)
	}
	if reads.Touched(changes) {
		t.Error("a read through a ref touched an unrelated change")
	}

	// Read the root World state.
	if err := ws.AccessWorldState(ctx, nil, noop); err != nil {
		t.Fatal(err)
	}
	if !reads.Touched(changes) {
		t.Error("a read of the root World state missed a change")
	}
}

// TestMatchGraphQuad matches a lookup filter against changed quads.
func TestMatchGraphQuad(t *testing.T) {
	filter := NewGraphQuadWithKeys("", "pred", "o", "")
	if !MatchGraphQuad(filter, NewGraphQuad("<s>", `"pred"`, "<o>", "")) {
		t.Error("bare predicate filter missed the stored quad")
	}
	if MatchGraphQuad(filter, NewGraphQuad("<s>", `"pred"`, "<o2>", "")) {
		t.Error("filter matched another object")
	}
}
