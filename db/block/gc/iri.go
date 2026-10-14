package block_gc

import (
	"strings"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/net/hash"
)

// IRI prefix constants for constructing graph node identifiers.
const (
	prefixBlock  = "block:"
	prefixObject = "object:"
	prefixBucket = "bucket:"
)

// Well-known node IRIs (permanent roots).
const (
	// NodeGCRoot owns everything reachable that no other root holds.
	NodeGCRoot = "gcroot"
	// NodeUnreferenced marks nodes that lost their last owner.
	NodeUnreferenced = "unreferenced"
	// NodeRefsKnown is the target of an edge from each block whose outgoing
	// edges were recorded when it was stored. A block without that edge may
	// hold refs the graph never saw, so its absence never means a leaf.
	NodeRefsKnown = "refs-known"
)

// PredGCRef is the single predicate for all GC reference edges.
const PredGCRef = "gc/ref"

// BlockIRI returns the IRI for a block reference: "block:" followed by the
// hash type byte, never zero, and the digest, whose length the type fixes.
// The IRI is binary; FormatIRI spells it for people.
func BlockIRI(ref *block.BlockRef) string {
	// An empty reference names no node.
	if ref == nil || ref.GetEmpty() {
		return ""
	}

	// Append the hash type byte and digest after the prefix.
	h := ref.GetHash()
	iri := make([]byte, 0, len(prefixBlock)+1+len(h.GetHash()))
	iri = append(iri, prefixBlock...)
	iri = append(iri, byte(h.GetHashType())) //nolint:gosec // hash types fit in a byte.
	iri = append(iri, h.GetHash()...)
	return string(iri)
}

// ParseBlockIRI parses a BlockIRI back to a BlockRef.
// Returns nil, false if not a valid block IRI.
func ParseBlockIRI(iri string) (*block.BlockRef, bool) {
	t, digest, ok := parseBlockIRI(iri)
	if !ok {
		return nil, false
	}
	return block.NewBlockRef(&hash.Hash{HashType: t, Hash: []byte(digest)}), true
}

// parseBlockIRI returns the hash type and digest of a block IRI, requiring
// a known type and a digest of its length.
func parseBlockIRI(iri string) (hash.HashType, string, bool) {
	// Require the block prefix and a nonempty suffix.
	rest, ok := strings.CutPrefix(iri, prefixBlock)
	if !ok || rest == "" {
		return 0, "", false
	}

	// Require a known hash type followed by a digest of its length.
	t := hash.HashType(rest[0])
	if size := t.GetHashLen(); size <= 0 || len(rest)-1 != size {
		return 0, "", false
	}
	return t, rest[1:], true
}

// FormatIRI returns iri for people: a block IRI spelled with the base58
// text of its reference, any other IRI unchanged.
func FormatIRI(iri string) string {
	if ref, ok := ParseBlockIRI(iri); ok {
		return prefixBlock + ref.MarshalString()
	}
	return iri
}

// LegacyBlockIRI converts a block IRI spelled with the base58 text of its
// reference, as earlier volumes stored it, to its BlockIRI. It returns any
// other IRI unchanged. Base58 text never begins with a byte that is a hash
// type, so the two forms cannot be confused.
func LegacyBlockIRI(iri string) string {
	// Decode the base58 reference after the block prefix.
	rest, ok := strings.CutPrefix(iri, prefixBlock)
	if !ok || rest == "" {
		return iri
	}
	ref, err := block.UnmarshalBlockRefB58(rest)
	if err != nil {
		return iri
	}
	return BlockIRI(ref)
}

// ObjectIRI returns "object:{key}" for a world object.
func ObjectIRI(key string) string {
	return prefixObject + key
}

// ParseObjectIRI parses an "object:{key}" IRI back to the object key.
// Returns the key and true if valid.
func ParseObjectIRI(iri string) (string, bool) {
	return strings.CutPrefix(iri, prefixObject)
}

// BucketIRI returns the IRI for a bucket node: "bucket:{bucketID}".
func BucketIRI(bucketID string) string {
	return prefixBucket + bucketID
}

// ParseBucketIRI parses a "bucket:{id}" IRI back to a bucket ID.
// Returns the bucket ID and true if valid.
func ParseBucketIRI(iri string) (string, bool) {
	rest, ok := strings.CutPrefix(iri, prefixBucket)
	if !ok || rest == "" {
		return "", false
	}
	return rest, true
}

// IsPermanentRoot returns true if the IRI is a well-known permanent root.
func IsPermanentRoot(iri string) bool {
	return iri == NodeGCRoot || iri == NodeUnreferenced || iri == NodeRefsKnown
}
