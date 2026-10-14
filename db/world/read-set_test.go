package world

import (
	"context"
	"testing"
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
