package s4db

import (
	"bytes"
	"cmp"
	"encoding/binary"

	"github.com/pkg/errors"
)

// Page layout: kind byte, entry count, entries, zero padding, and a CRC-32C of
// the rest in the last four bytes.
const (
	// pageHeader is the kind byte and the two-byte count.
	pageHeader = 3
	// pageRoom is the space for entries.
	pageRoom = pageSize - pageHeader - 4
	// maxInline is the largest inline value; a leaf entry holding it and
	// the longest key fits in a page.
	maxInline = 1024
)

// Page kinds.
const (
	// pageInner holds keys and child pages.
	pageInner = 0
	// pageLeaf holds keys and values.
	pageLeaf = 1
)

// Decoded entry costs in memory on a 64-bit platform, beyond the page bytes
// the node keeps.
const (
	// leafEntryCost is an entry offset and its head.
	leafEntryCost = 2 + 8
	// innerEntryCost is an entry offset and its head.
	innerEntryCost = 2 + 8
)

// node is a decoded tree page. It keeps the page with the offset and head of
// each entry, and reads keys, values, and child pages from the page when
// asked, so a cached page holds no pointer per entry.
type node struct {
	// leaf is set for leaf pages.
	leaf bool
	// page is the encoded page.
	page []byte
	// offs holds the offset of each entry in page.
	offs []uint16
	// pre is the prefix every searched key shares: each key of a leaf, and
	// each key after the empty first one of an inner page. It points into
	// page.
	pre []byte
	// heads holds the head of each key after pre, so a search compares
	// within one array and reads a key only on a tie. Keys of one kind
	// share a long prefix, such as the type bytes of a block reference, and
	// heads taken before it would all tie. The empty first inner key has
	// head zero and is never searched.
	heads []uint64
	// cost is the memory the decoded node holds, for the cache budget.
	cost int
}

// head returns the first eight bytes of key as a big-endian integer, zero
// padded, so integers order as their keys do except on ties.
func head(key []byte) uint64 {
	if len(key) >= 8 {
		return binary.BigEndian.Uint64(key)
	}
	var b [8]byte
	copy(b[:], key)
	return binary.BigEndian.Uint64(b[:])
}

// count returns the number of entries.
func (n *node) count() int {
	return len(n.offs)
}

// entry returns a decoder at entry i. Decoding checked every entry, so
// reading one again cannot fail.
func (n *node) entry(i int) decoder {
	return decoder{b: n.page[n.offs[i] : pageSize-4]}
}

// key returns the key of entry i, a leaf key or the low key of a child.
// The first inner key is empty.
func (n *node) key(i int) []byte {
	d := n.entry(i)
	return d.bytes()
}

// val returns the value of leaf entry i.
func (n *node) val(i int) value {
	d := n.entry(i)
	d.bytes()
	return d.value(d.kind())
}

// kid returns the child page of inner entry i.
func (n *node) kid(i int) uint64 {
	d := n.entry(i)
	d.bytes()
	return d.u64()
}

// search returns the position of the first key at or after key in leaf n
// and whether it equals key.
func (n *node) search(key []byte) (int, bool) {
	return n.searchFrom(0, key)
}

// searchFrom is search over the keys from position lo, which must not
// include the empty first key of an inner page.
func (n *node) searchFrom(lo int, key []byte) (int, bool) {
	// Every key from lo starts with pre, so a key without it sorts before
	// or after all of them.
	p := len(n.pre)
	if len(key) < p || !bytes.Equal(key[:p], n.pre) {
		if bytes.Compare(key, n.pre) < 0 {
			return lo, false
		}
		return len(n.heads), false
	}

	// Bisect on heads, comparing whole keys on equal heads.
	h := head(key[p:])
	hi := len(n.heads)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		c := cmp.Compare(n.heads[m], h)
		if c == 0 {
			c = bytes.Compare(n.key(m), key)
		}
		if c < 0 {
			lo = m + 1
			continue
		}
		hi = m
	}
	return lo, lo < len(n.heads) && n.heads[lo] == h && bytes.Equal(n.key(lo), key)
}

// childIndex returns the child of inner page n covering key: the last child
// whose low key is at or before key. The first child covers everything below
// the second, so its key is never compared.
func (n *node) childIndex(key []byte) int {
	i, found := n.searchFrom(1, key)
	if found {
		return i
	}
	return i - 1
}

// fill returns the bytes the entries of n occupy.
func (n *node) fill() int {
	// The entries run from the header to the end of the last one.
	if len(n.offs) == 0 {
		return 0
	}
	d := n.entry(len(n.offs) - 1)
	d.bytes()
	if n.leaf {
		d.value(d.kind())
	} else {
		d.u64()
	}
	return pageSize - 4 - len(d.b) - pageHeader
}

// leafEntrySize returns the encoded length of a leaf entry.
func leafEntrySize(key []byte, v value) int {
	return uvarintLen(uint64(len(key))) + len(key) + v.size()
}

// innerEntrySize returns the encoded length of an inner entry.
func innerEntrySize(key []byte) int {
	return uvarintLen(uint64(len(key))) + len(key) + 8
}

// decodeNode reads a page, keeping b.
func decodeNode(b []byte) (*node, error) {
	// Check the page checksum and kind.
	if checksum(b[:pageSize-4]) != binary.LittleEndian.Uint32(b[pageSize-4:]) {
		return nil, errors.Wrap(ErrCorrupt, "page checksum mismatch")
	}
	if b[0] != pageInner && b[0] != pageLeaf {
		return nil, errors.Wrap(ErrCorrupt, "unknown page kind")
	}

	// Size the node for its kind and count.
	count := int(binary.LittleEndian.Uint16(b[1:]))
	n := &node{leaf: b[0] == pageLeaf, page: b, offs: make([]uint16, count), heads: make([]uint64, count)}
	d := decoder{b: b[pageHeader : pageSize-4]}
	n.cost = pageSize + count*innerEntryCost
	if n.leaf {
		n.cost = pageSize + count*leafEntryCost
	}

	// Find each entry, checking that its value or child page is whole.
	// Until the prefix is known, each head holds where its key ends and
	// how long it is, so the heads need no second decode.
	for i := range count {
		n.offs[i] = uint16(pageSize - 4 - len(d.b)) // #nosec G115 -- offsets fall within a page.
		k := d.bytes()
		n.heads[i] = uint64(pageSize-4-len(d.b))<<16 | uint64(len(k)) // #nosec G115 -- ends and lengths fall within a page.
		if n.leaf {
			d.value(d.kind())
			continue
		}
		d.u64()
	}
	if d.err != nil {
		return nil, d.err
	}

	// Take the prefix the searched keys share. They are sorted, so it is
	// the common prefix of the first and the last.
	key := func(i int) []byte {
		end, l := int(n.heads[i]>>16), int(n.heads[i]&0xffff)
		return b[end-l : end]
	}
	first := 0
	if !n.leaf {
		first = 1
		n.heads[0] = 0
	}
	if first < count {
		lo, hi := key(first), key(count-1)
		p := 0
		for p < len(lo) && p < len(hi) && lo[p] == hi[p] {
			p++
		}
		n.pre = lo[:p:p]
	}

	// Take the head of each searched key after the prefix.
	for i := first; i < count; i++ {
		n.heads[i] = head(key(i)[len(n.pre):])
	}
	return n, nil
}
