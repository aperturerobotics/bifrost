package segment

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"strconv"

	"github.com/pkg/errors"
	trace "github.com/s4wave/spacewave/db/traceutil"
)

// maxExistenceWindowRead keeps small existence lookups to one buffered read;
// larger windows stream entry metadata to avoid copying their values.
const maxExistenceWindowRead = 64 * 1024

// maxLookupWindowRead keeps value lookups through 256 KiB to one buffered
// read; larger windows stream to bound allocation.
const maxLookupWindowRead = 256 * 1024

// LookupMeta is the metadata needed for point lookups without reparsing the
// full SSTable on each access.
type LookupMeta struct {
	// Header describes the encoded segment sections.
	Header *Header
	// MinKey is the first key in the segment.
	MinKey []byte
	// MaxKey is the last key in the segment.
	MaxKey []byte
	// Index maps sparse keys to data-window offsets.
	Index []IndexEntry
	// Bloom rejects keys that cannot occur in the segment.
	Bloom *BloomFilter
}

// LookupResult is the result of a batched segment lookup.
type LookupResult struct {
	// Value contains a copied live value when the lookup requested values.
	Value []byte
	// Found reports whether the segment contains a live value.
	Found bool
	// Tombstone reports whether the segment contains a deletion marker.
	Tombstone bool
}

// LookupStat is metadata for a lookup result that does not require loading the
// value bytes.
type LookupStat struct {
	// ValueSize is the encoded live value length.
	ValueSize int64
	// Found reports whether the segment contains a live value.
	Found bool
	// Tombstone reports whether the segment contains a deletion marker.
	Tombstone bool
}

// LoadLookupMeta loads only the SSTable metadata needed for point lookups.
func LoadLookupMeta(r io.ReaderAt, size int64) (*LookupMeta, error) {
	// Decode the fixed header after checking the minimum encoded size.
	if size < HeaderSize+4 {
		return nil, errors.New("file too small for SSTable")
	}

	var hdrBuf [HeaderSize]byte
	if _, err := r.ReadAt(hdrBuf[:], 0); err != nil {
		return nil, errors.Wrap(err, "read header")
	}
	hdr, err := DecodeHeader(hdrBuf[:])
	if err != nil {
		return nil, errors.Wrap(err, "decode header")
	}

	// Decode the key bounds that guard every lookup.
	keyMetaSize := 2 + int(hdr.MinKeySize) + 2 + int(hdr.MaxKeySize)
	keyBuf := make([]byte, keyMetaSize)
	if _, err := r.ReadAt(keyBuf, HeaderSize); err != nil {
		return nil, errors.Wrap(err, "read key metadata")
	}

	off := 0
	minKeyLen := int(binary.BigEndian.Uint16(keyBuf[off : off+2]))
	off += 2
	if off+minKeyLen > len(keyBuf) {
		return nil, errors.New("truncated min key")
	}
	minKey := make([]byte, minKeyLen)
	copy(minKey, keyBuf[off:off+minKeyLen])
	off += minKeyLen

	maxKeyLen := int(binary.BigEndian.Uint16(keyBuf[off : off+2]))
	off += 2
	if off+maxKeyLen > len(keyBuf) {
		return nil, errors.New("truncated max key")
	}
	maxKey := make([]byte, maxKeyLen)
	copy(maxKey, keyBuf[off:off+maxKeyLen])

	// Decode the optional sparse index.
	var idx []IndexEntry
	if hdr.IndexSize > 0 {
		idxBuf := make([]byte, hdr.IndexSize)
		if _, err := r.ReadAt(idxBuf, int64(hdr.IndexOffset)); err != nil {
			return nil, errors.Wrap(err, "read index block")
		}
		idx, err = decodeIndex(idxBuf)
		if err != nil {
			return nil, errors.Wrap(err, "decode index")
		}
	}

	// Decode the optional negative-lookup filter.
	var bloom *BloomFilter
	if hdr.BloomSize > 0 {
		bloomBuf := make([]byte, hdr.BloomSize)
		if _, err := r.ReadAt(bloomBuf, int64(hdr.BloomOffset)); err != nil {
			return nil, errors.Wrap(err, "read bloom block")
		}
		bloom, err = DecodeBloom(bloomBuf)
		if err != nil {
			return nil, errors.Wrap(err, "decode bloom")
		}
	}

	// Retain only metadata needed by point and batch lookups.
	return &LookupMeta{
		Header: hdr,
		MinKey: minKey,
		MaxKey: maxKey,
		Index:  idx,
		Bloom:  bloom,
	}, nil
}

// Get looks up a live value using cached segment metadata.
func (m *LookupMeta) Get(r io.ReaderAt, key []byte) ([]byte, bool, error) {
	val, found, _, err := m.Locate(r, key, true)
	return val, found, err
}

// Has checks whether a key exists without materializing a result value.
// Tombstoned keys return false.
func (m *LookupMeta) Has(r io.ReaderAt, key []byte) (bool, error) {
	_, found, _, err := m.Locate(r, key, false)
	return found, err
}

// Stat resolves a key and returns the value size without materializing the
// value. Tombstoned keys return Found=false with Tombstone=true.
func (m *LookupMeta) Stat(r io.ReaderAt, key []byte) (LookupStat, error) {
	// Trace the complete metadata lookup.
	ctx := context.Background()
	_, task := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/stat")
	defer task.End()

	// Reject keys excluded by the segment bounds or bloom filter.
	keyStr := string(key)
	if keyStr < string(m.MinKey) || keyStr > string(m.MaxKey) {
		return LookupStat{}, nil
	}
	if m.Bloom != nil && !m.Bloom.MayContain(key) {
		return LookupStat{}, nil
	}

	// Select and validate the sparse-index window.
	start, limit := SearchIndex(m.Index, key, m.Header.DataSize)
	if limit < start {
		return LookupStat{}, errors.New("invalid data window")
	}
	windowSize, err := uint32ToInt(limit - start)
	if err != nil {
		return LookupStat{}, err
	}

	// Buffer common windows and stream larger windows to bound allocation.
	if windowSize <= maxLookupWindowRead {
		window := make([]byte, windowSize)
		if _, err := r.ReadAt(window, int64(m.Header.DataOffset)+int64(start)); err != nil {
			return LookupStat{}, errors.Wrap(err, "read data window")
		}
		return statInWindowBytes(window, keyStr)
	}
	return statInWindowReader(r, int64(m.Header.DataOffset), start, limit, key)
}

// Locate resolves a key using cached metadata.
// Returns either a live value, a tombstone marker, or a miss.
func (m *LookupMeta) Locate(r io.ReaderAt, key []byte, loadValue bool) ([]byte, bool, bool, error) {
	// Trace the complete point lookup.
	ctx := context.Background()
	ctx, task := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate")
	defer task.End()

	// Reject keys excluded by the segment bounds or bloom filter.
	keyStr := string(key)
	if keyStr < string(m.MinKey) || keyStr > string(m.MaxKey) {
		return nil, false, false, nil
	}
	if m.Bloom != nil && !m.Bloom.MayContain(key) {
		return nil, false, false, nil
	}

	// Select and validate the sparse-index window.
	_, subtask := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate/search-index")
	start, limit := SearchIndex(m.Index, key, m.Header.DataSize)
	subtask.End()

	if limit < start {
		return nil, false, false, errors.New("invalid data window")
	}
	windowSize, err := uint32ToInt(limit - start)
	if err != nil {
		return nil, false, false, err
	}
	trace.Log(ctx, "window", "size="+strconv.Itoa(windowSize))

	// Stream large windows so existence checks avoid copying value bytes and
	// value lookups bound their allocation.
	if (!loadValue && windowSize > maxExistenceWindowRead) || windowSize > maxLookupWindowRead {
		_, subtask = trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate/scan-window-streamed")
		val, found, tombstone, err := locateInWindowReader(r, int64(m.Header.DataOffset), start, limit, key, loadValue)
		subtask.End()
		return val, found, tombstone, err
	}

	// Read a common window once before scanning it in memory.
	_, subtask = trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate/read-window")
	window := make([]byte, windowSize)
	if _, err := r.ReadAt(window, int64(m.Header.DataOffset)+int64(start)); err != nil {
		subtask.End()
		return nil, false, false, errors.Wrap(err, "read data window")
	}
	subtask.End()

	// Resolve the requested key from the buffered window.
	taskCtx, subtask := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate/scan-window")
	val, found, tombstone, err := locateInWindowBytes(taskCtx, window, keyStr, loadValue)
	subtask.End()
	return val, found, tombstone, err
}

// locateInWindowBytes searches one buffered sparse-index window.
func locateInWindowBytes(
	ctx context.Context,
	window []byte,
	keyStr string,
	loadValue bool,
) ([]byte, bool, bool, error) {
	// Scan ordered entries until the key is found or passed.
	off := 0
	for off < len(window) {
		// Decode the next entry key.
		if off+2 > len(window) {
			break
		}
		keyLen := int(binary.BigEndian.Uint16(window[off : off+2]))
		off += 2
		if off+keyLen > len(window) {
			break
		}
		entryKey := string(window[off : off+keyLen])
		off += keyLen

		// Decode the value marker and length.
		if off+4 > len(window) {
			break
		}
		valLen := binary.BigEndian.Uint32(window[off : off+4])
		off += 4

		// Return the matching live value or tombstone state.
		if entryKey == keyStr {
			if valLen == TombstoneLen {
				return nil, false, true, nil
			}
			if !loadValue {
				return nil, true, false, nil
			}
			if uint64(len(window)-off) < uint64(valLen) {
				return nil, false, false, errors.New("truncated value in data window")
			}
			valLenInt, err := uint32ToInt(valLen)
			if err != nil {
				return nil, false, false, err
			}
			_, copyTask := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate/copy-value")
			val := make([]byte, valLenInt)
			copy(val, window[off:off+valLenInt])
			copyTask.End()
			return val, true, false, nil
		}
		if entryKey > keyStr {
			return nil, false, false, nil
		}

		// Advance past an unmatched live value without copying it.
		if valLen != TombstoneLen {
			valLenInt, err := uint32ToInt(valLen)
			if err != nil {
				return nil, false, false, err
			}
			if len(window)-off < valLenInt {
				break
			}
			off += valLenInt
		}
	}
	return nil, false, false, nil
}

// locateInWindowReader searches one sparse-index window through bounded reads.
func locateInWindowReader(
	r io.ReaderAt,
	dataOffset int64,
	start uint32,
	limit uint32,
	key []byte,
	loadValue bool,
) ([]byte, bool, bool, error) {
	// Scan ordered entry headers until the key is found or passed.
	off := start
	var header [4]byte
	for off < limit {
		// Read the next entry key through bounded requests.
		if limit-off < 2 {
			break
		}
		if _, err := r.ReadAt(header[:2], dataOffset+int64(off)); err != nil {
			return nil, false, false, errors.Wrap(err, "read data window key length")
		}
		keyLen := uint32(binary.BigEndian.Uint16(header[:2]))
		off += 2
		if limit-off < keyLen {
			break
		}

		entryKey := make([]byte, keyLen)
		if keyLen != 0 {
			if _, err := r.ReadAt(entryKey, dataOffset+int64(off)); err != nil {
				return nil, false, false, errors.Wrap(err, "read data window key")
			}
		}
		off += keyLen

		// Read the value marker and length.
		if limit-off < 4 {
			break
		}
		if _, err := r.ReadAt(header[:4], dataOffset+int64(off)); err != nil {
			return nil, false, false, errors.Wrap(err, "read data window value length")
		}
		valLen := binary.BigEndian.Uint32(header[:4])
		off += 4

		// Return the matching live value or tombstone state.
		cmp := bytes.Compare(entryKey, key)
		if cmp == 0 {
			if valLen == TombstoneLen {
				return nil, false, true, nil
			}
			if !loadValue {
				return nil, true, false, nil
			}
			if uint64(limit-off) < uint64(valLen) {
				return nil, false, false, errors.New("truncated value in data window")
			}
			valLenInt, err := uint32ToInt(valLen)
			if err != nil {
				return nil, false, false, err
			}
			val := make([]byte, valLenInt)
			if valLenInt != 0 {
				if _, err := r.ReadAt(val, dataOffset+int64(off)); err != nil {
					return nil, false, false, errors.Wrap(err, "read data window value")
				}
			}
			return val, true, false, nil
		}
		if cmp > 0 {
			return nil, false, false, nil
		}

		// Advance past an unmatched live value without reading it.
		if valLen == TombstoneLen {
			continue
		}
		if uint64(limit-off) < uint64(valLen) {
			break
		}
		off += valLen
	}
	return nil, false, false, nil
}

// statInWindowBytes resolves value metadata in one buffered sparse-index window.
func statInWindowBytes(window []byte, keyStr string) (LookupStat, error) {
	// Scan ordered entries until the key is found or passed.
	off := 0
	for off < len(window) {
		// Decode the next entry key.
		if off+2 > len(window) {
			break
		}
		keyLen := int(binary.BigEndian.Uint16(window[off : off+2]))
		off += 2
		if off+keyLen > len(window) {
			break
		}
		entryKey := string(window[off : off+keyLen])
		off += keyLen

		// Decode the value marker and length.
		if off+4 > len(window) {
			break
		}
		valLen := binary.BigEndian.Uint32(window[off : off+4])
		off += 4

		// Return metadata for the matching live value or tombstone.
		if entryKey == keyStr {
			if valLen == TombstoneLen {
				return LookupStat{Tombstone: true}, nil
			}
			if uint64(len(window)-off) < uint64(valLen) {
				return LookupStat{}, errors.New("truncated value in data window")
			}
			return LookupStat{ValueSize: int64(valLen), Found: true}, nil
		}
		if entryKey > keyStr {
			return LookupStat{}, nil
		}

		// Advance past an unmatched live value.
		if valLen != TombstoneLen {
			valLenInt, err := uint32ToInt(valLen)
			if err != nil {
				return LookupStat{}, err
			}
			if len(window)-off < valLenInt {
				break
			}
			off += valLenInt
		}
	}
	return LookupStat{}, nil
}

// statInWindowReader resolves value metadata through bounded reads.
func statInWindowReader(
	r io.ReaderAt,
	dataOffset int64,
	start uint32,
	limit uint32,
	key []byte,
) (LookupStat, error) {
	// Scan ordered entry headers until the key is found or passed.
	off := start
	var header [4]byte
	for off < limit {
		// Read the next entry key through bounded requests.
		if limit-off < 2 {
			break
		}
		if _, err := r.ReadAt(header[:2], dataOffset+int64(off)); err != nil {
			return LookupStat{}, errors.Wrap(err, "read data window key length")
		}
		keyLen := uint32(binary.BigEndian.Uint16(header[:2]))
		off += 2
		if limit-off < keyLen {
			break
		}

		entryKey := make([]byte, keyLen)
		if keyLen != 0 {
			if _, err := r.ReadAt(entryKey, dataOffset+int64(off)); err != nil {
				return LookupStat{}, errors.Wrap(err, "read data window key")
			}
		}
		off += keyLen

		// Read the value marker and length.
		if limit-off < 4 {
			break
		}
		if _, err := r.ReadAt(header[:4], dataOffset+int64(off)); err != nil {
			return LookupStat{}, errors.Wrap(err, "read data window value length")
		}
		valLen := binary.BigEndian.Uint32(header[:4])
		off += 4

		// Return metadata for the matching live value or tombstone.
		cmp := bytes.Compare(entryKey, key)
		if cmp == 0 {
			if valLen == TombstoneLen {
				return LookupStat{Tombstone: true}, nil
			}
			if uint64(limit-off) < uint64(valLen) {
				return LookupStat{}, errors.New("truncated value in data window")
			}
			return LookupStat{ValueSize: int64(valLen), Found: true}, nil
		}
		if cmp > 0 {
			return LookupStat{}, nil
		}

		// Advance past an unmatched live value without reading it.
		if valLen == TombstoneLen {
			continue
		}
		if uint64(limit-off) < uint64(valLen) {
			break
		}
		off += valLen
	}
	return LookupStat{}, nil
}

// LocateBatch resolves keys using cached metadata and groups keys by
// sparse-index window.
func (m *LookupMeta) LocateBatch(r io.ReaderAt, keys [][]byte, loadValue bool) ([]LookupResult, error) {
	// Trace the complete batch lookup.
	ctx := context.Background()
	ctx, task := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate-batch")
	defer task.End()

	// Preserve input order and return immediately for an empty batch.
	out := make([]LookupResult, len(keys))
	if len(keys) == 0 {
		return out, nil
	}

	// Group eligible keys by sparse-index window.
	type lookupWindow struct {
		// start is the inclusive data offset.
		start uint32
		// limit is the exclusive data offset.
		limit uint32
		// keys indexes requested keys assigned to the window.
		keys []int
	}
	var windows []lookupWindow
	for i, key := range keys {
		keyStr := string(key)
		if keyStr < string(m.MinKey) || keyStr > string(m.MaxKey) {
			continue
		}
		if m.Bloom != nil && !m.Bloom.MayContain(key) {
			continue
		}

		start, limit := SearchIndex(m.Index, key, m.Header.DataSize)
		var found bool
		for j := range windows {
			if windows[j].start == start && windows[j].limit == limit {
				windows[j].keys = append(windows[j].keys, i)
				found = true
				break
			}
		}
		if !found {
			windows = append(windows, lookupWindow{
				start: start,
				limit: limit,
				keys:  []int{i},
			})
		}
	}

	// Resolve each grouped window through its bounded read strategy.
	for _, lw := range windows {
		// Validate the encoded window bounds.
		if lw.limit < lw.start {
			return nil, errors.New("invalid data window")
		}
		windowSize, err := uint32ToInt(lw.limit - lw.start)
		if err != nil {
			return nil, err
		}
		trace.Log(ctx, "window", "size="+strconv.Itoa(windowSize))

		// Map duplicate requested keys back to every result position.
		want := make(map[string][]int, len(lw.keys))
		for _, keyIdx := range lw.keys {
			keyStr := string(keys[keyIdx])
			want[keyStr] = append(want[keyStr], keyIdx)
		}

		// Resolve common windows from one buffered read.
		if windowSize <= maxLookupWindowRead {
			_, subtask := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate-batch/read-window")
			window := make([]byte, windowSize)
			if _, err := r.ReadAt(window, int64(m.Header.DataOffset)+int64(lw.start)); err != nil {
				subtask.End()
				return nil, errors.Wrap(err, "read data window")
			}
			subtask.End()

			_, subtask = trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate-batch/scan-window")
			if err := locateBatchInWindowBytes(window, want, out, loadValue); err != nil {
				subtask.End()
				return nil, err
			}
			subtask.End()
			continue
		}

		// Stream larger windows to bound allocation.
		_, subtask := trace.NewTask(ctx, "hydra/opfs-segment/lookup-meta/locate-batch/scan-window-streamed")
		if err := locateBatchInWindowReader(r, int64(m.Header.DataOffset), lw.start, lw.limit, want, out, loadValue); err != nil {
			subtask.End()
			return nil, err
		}
		subtask.End()
	}

	return out, nil
}

// locateBatchInWindowBytes resolves a key group in one buffered window.
func locateBatchInWindowBytes(
	window []byte,
	want map[string][]int,
	out []LookupResult,
	loadValue bool,
) error {
	// Scan ordered entries until every requested key is resolved.
	off := 0
	for off < len(window) && len(want) != 0 {
		// Decode the next entry key.
		if off+2 > len(window) {
			break
		}
		keyLen := int(binary.BigEndian.Uint16(window[off : off+2]))
		off += 2
		if off+keyLen > len(window) {
			break
		}
		entryKey := string(window[off : off+keyLen])
		off += keyLen

		// Decode the value marker and length.
		if off+4 > len(window) {
			break
		}
		valLen := binary.BigEndian.Uint32(window[off : off+4])
		off += 4

		// Populate every result requested for the matching key.
		if keyIdxs, ok := want[entryKey]; ok {
			if valLen == TombstoneLen {
				for _, keyIdx := range keyIdxs {
					out[keyIdx].Tombstone = true
				}
				delete(want, entryKey)
				continue
			}
			if loadValue {
				if uint64(len(window)-off) < uint64(valLen) {
					return errors.New("truncated value in data window")
				}
				valLenInt, err := uint32ToInt(valLen)
				if err != nil {
					return err
				}
				for _, keyIdx := range keyIdxs {
					val := make([]byte, valLenInt)
					copy(val, window[off:off+valLenInt])
					out[keyIdx].Value = val
					out[keyIdx].Found = true
				}
			} else {
				for _, keyIdx := range keyIdxs {
					out[keyIdx].Found = true
				}
			}
			delete(want, entryKey)
		}

		// Advance to the next entry after validating its encoded value.
		if valLen != TombstoneLen {
			valLenInt, err := uint32ToInt(valLen)
			if err != nil {
				return err
			}
			if len(window)-off < valLenInt {
				break
			}
			off += valLenInt
		}
	}
	return nil
}

// locateBatchInWindowReader resolves a key group through bounded reads.
func locateBatchInWindowReader(
	r io.ReaderAt,
	dataOffset int64,
	start uint32,
	limit uint32,
	want map[string][]int,
	out []LookupResult,
	loadValue bool,
) error {
	// Scan ordered entry headers until every requested key is resolved.
	off := start
	var header [4]byte
	for off < limit && len(want) != 0 {
		// Read the next entry key through bounded requests.
		if limit-off < 2 {
			break
		}
		if _, err := r.ReadAt(header[:2], dataOffset+int64(off)); err != nil {
			return errors.Wrap(err, "read data window key length")
		}
		keyLen := uint32(binary.BigEndian.Uint16(header[:2]))
		off += 2
		if limit-off < keyLen {
			break
		}

		entryKey := make([]byte, keyLen)
		if keyLen != 0 {
			if _, err := r.ReadAt(entryKey, dataOffset+int64(off)); err != nil {
				return errors.Wrap(err, "read data window key")
			}
		}
		off += keyLen

		// Read the value marker and length.
		if limit-off < 4 {
			break
		}
		if _, err := r.ReadAt(header[:4], dataOffset+int64(off)); err != nil {
			return errors.Wrap(err, "read data window value length")
		}
		valLen := binary.BigEndian.Uint32(header[:4])
		off += 4

		// Populate every result requested for the matching key.
		entryKeyStr := string(entryKey)
		if keyIdxs, ok := want[entryKeyStr]; ok {
			if valLen == TombstoneLen {
				for _, keyIdx := range keyIdxs {
					out[keyIdx].Tombstone = true
				}
				delete(want, entryKeyStr)
				continue
			}
			if loadValue {
				if uint64(limit-off) < uint64(valLen) {
					return errors.New("truncated value in data window")
				}
				valLenInt, err := uint32ToInt(valLen)
				if err != nil {
					return err
				}
				for _, keyIdx := range keyIdxs {
					val := make([]byte, valLenInt)
					if valLenInt != 0 {
						if _, err := r.ReadAt(val, dataOffset+int64(off)); err != nil {
							return errors.Wrap(err, "read data window value")
						}
					}
					out[keyIdx].Value = val
					out[keyIdx].Found = true
				}
			} else {
				for _, keyIdx := range keyIdxs {
					out[keyIdx].Found = true
				}
			}
			delete(want, entryKeyStr)
		}

		// Advance to the next entry after validating its encoded value.
		if valLen == TombstoneLen {
			continue
		}
		if uint64(limit-off) < uint64(valLen) {
			break
		}
		off += valLen
	}
	return nil
}

// uint32ToInt rejects encoded lengths that the host cannot represent.
func uint32ToInt(v uint32) (int, error) {
	if uint64(v) > uint64(int(^uint(0)>>1)) {
		return 0, errors.New("value length exceeds maximum")
	}
	return int(v), nil
}
