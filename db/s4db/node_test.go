package s4db

import (
	"bytes"
	"slices"
	"testing"
)

// TestNodeSearchSharedPrefix checks leaf search and inner routing on pages
// whose keys share a long prefix, as block keys do, against a sorted key
// list, for probes inside the prefix, before it, after it, and shorter
// than it.
func TestNodeSearchSharedPrefix(t *testing.T) {
	// Build a leaf and an inner page over keys sharing a ten-byte prefix.
	pre := []byte("h/b/\x0a\x24\x08\x01\x12\x20")
	var keys [][]byte
	for _, tail := range []string{"", "\x00", "\x01\x00", "abcdefgh1", "abcdefgh2", "abcdefgh2\x00", "z"} {
		keys = append(keys, append(bytes.Clone(pre), tail...))
	}
	leafDraft := &draft{leaf: true, keys: keys, vals: make([]value, len(keys))}
	leaf, err := decodeNode(leafDraft.encode(nil))
	if err != nil {
		t.Fatal(err)
	}
	innerDraft := &draft{keys: append([][]byte{{}}, keys...), kids: make([]uint64, len(keys)+1)}
	for i := range innerDraft.kids {
		innerDraft.kids[i] = uint64(i) + 10
	}
	inner, err := decodeNode(innerDraft.encode(nil))
	if err != nil {
		t.Fatal(err)
	}

	// Probe every key, its neighbors, and keys outside the prefix.
	probes := [][]byte{{}, []byte("h"), []byte("h/b/"), pre[:9], []byte("h/b/\x0a\x25"), []byte("i")}
	for _, k := range keys {
		probes = append(probes, k, append(bytes.Clone(k), 0), k[:len(k)-1])
	}
	for _, p := range probes {
		want, wantFound := slices.BinarySearchFunc(keys, p, bytes.Compare)
		if got, found := leaf.search(p); got != want || found != wantFound {
			t.Errorf("leaf search %q = %d, %v, want %d, %v", p, got, found, want, wantFound)
		}

		// The inner page routes to the last child whose low key is at or
		// before the probe; its first child covers everything below.
		wantKid := want
		if wantFound {
			wantKid++
		}
		if got := inner.childIndex(p); got != wantKid {
			t.Errorf("inner childIndex %q = %d, want %d", p, got, wantKid)
		}
	}
}
