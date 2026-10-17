package paylog

import (
	"context"
	"encoding/binary"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/db/volume/device"
)

// moveBytes bounds the live payload bytes one maintenance pass moves.
const moveBytes = 8 << 20

// segmentStats counts the published blocks in a segment.
type segmentStats struct {
	// count is the number of blocks.
	count int64
	// bytes is the total payload length.
	bytes int64
}

// add returns the statistics with count blocks of bytes total added.
func (st segmentStats) add(count, bytes int64) segmentStats {
	return segmentStats{count: st.count + count, bytes: st.bytes + bytes}
}

// marshal encodes the statistics as two unsigned varints.
func (st segmentStats) marshal() []byte {
	b := binary.AppendUvarint(nil, uint64(st.count)) //nolint:gosec
	return binary.AppendUvarint(b, uint64(st.bytes)) //nolint:gosec
}

// unmarshalSegmentStats decodes segment statistics.
func unmarshalSegmentStats(b []byte) (segmentStats, error) {
	var v [2]uint64
	for i := range v {
		n, size := binary.Uvarint(b)
		if size <= 0 || n > 1<<62 {
			return segmentStats{}, errors.New("invalid segment statistics")
		}
		v[i], b = n, b[size:]
	}
	return segmentStats{count: int64(v[0]), bytes: int64(v[1])}, nil //nolint:gosec
}

// segmentStatsKey returns the index key of segment n's statistics.
func segmentStatsKey(n uint64) []byte {
	return binary.BigEndian.AppendUint64([]byte(segmentStatsPrefix), n)
}

// addSegmentStats adds delta to segment n's statistics in tx, removing them
// when no block is left.
func addSegmentStats(ctx context.Context, tx kvtx.Tx, n uint64, delta segmentStats) error {
	// Read the current statistics.
	key := segmentStatsKey(n)
	var st segmentStats
	v, found, err := tx.Get(ctx, key)
	if err != nil {
		return err
	}
	if found {
		if st, err = unmarshalSegmentStats(v); err != nil {
			return err
		}
	}

	// Write the sum, or remove it with the segment's last block.
	st = st.add(delta.count, delta.bytes)
	if st.count <= 0 {
		return tx.Delete(ctx, key)
	}
	return tx.Set(ctx, key, st.marshal())
}

// readSegmentStats reads the statistics of every segment with a published
// block from tx.
func readSegmentStats(ctx context.Context, tx kvtx.Tx) (map[uint64]segmentStats, error) {
	stats := make(map[uint64]segmentStats)
	err := tx.ScanPrefix(ctx, []byte(segmentStatsPrefix), func(key, value []byte) error {
		st, err := unmarshalSegmentStats(value)
		if err != nil {
			return err
		}
		stats[binary.BigEndian.Uint64(key[len(segmentStatsPrefix):])] = st
		return nil
	})
	return stats, err
}

// BlockStats returns the number and total payload length of the published
// blocks.
func (s *Store) BlockStats(ctx context.Context) (count, size uint64, err error) {
	var stats map[uint64]segmentStats
	err = s.view(ctx, func(tx kvtx.Tx) error {
		stats, err = readSegmentStats(ctx, tx)
		return err
	})
	for _, st := range stats {
		count += uint64(st.count) //nolint:gosec
		size += uint64(st.bytes)  //nolint:gosec
	}
	return count, size, err
}

// Maintenance reclaims segment space in one bounded pass. It removes the
// sealed segments that hold no published or pending block, then moves up to
// moveBytes of live blocks out of the sealed segment with the most dead bytes,
// if at least half of it is dead. A later pass removes the emptied segment.
//
// The pass holds the writer only to read the index and remove segments, so no
// write transaction holds unpublished blocks in them meanwhile, and to publish
// the moves. It copies payloads without the writer. It takes one call at a
// time, since only it removes segments.
func (s *Store) Maintenance(ctx context.Context) error {
	// Remove the empty segments and choose the blocks to move.
	victim, keys, locs, err := s.sweepSegments(ctx)
	if err != nil || len(keys) == 0 {
		return err
	}

	// Copy them to the end of the log and publish their new locations.
	return s.move(ctx, victim, keys, locs)
}

// sweepSegments removes the empty sealed segments under the writer and
// returns up to moveBytes of the blocks in the sealed segment with the most
// dead bytes, if at least half of it is dead.
func (s *Store) sweepSegments(ctx context.Context) (victim uint64, keys []string, locs []*location, err error) {
	// Take the writer for the index reads and removes.
	t, err := s.begin(ctx, true, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	defer t.Discard()

	// Snapshot the open segment and the segments with pending payloads.
	s.mtx.Lock()
	open := s.segment
	busy := make(map[uint64]bool)
	for _, p := range s.pending {
		if p.loc != nil {
			busy[p.loc.segment] = true
		}
	}
	s.mtx.Unlock()

	// Read the published statistics and the segment files.
	stats, err := readSegmentStats(ctx, t.itx)
	if err != nil {
		return 0, nil, nil, err
	}
	files, err := s.dev.List(ctx)
	if err != nil {
		return 0, nil, nil, err
	}

	// Find the empty sealed segments and the one with the most dead bytes.
	var empty []string
	victimDead := int64(-1)
	for _, f := range files {
		n, ok := parseSegment(f.Name)
		if !ok || n >= open || busy[n] {
			continue
		}
		st := stats[n]
		if st.count == 0 {
			empty = append(empty, f.Name)
			continue
		}
		if dead := f.Size - st.bytes; dead*2 >= f.Size && dead > victimDead {
			victim, victimDead = n, dead
		}
	}

	// Remove the empty segments.
	if len(empty) != 0 {
		if err := s.dev.Remove(ctx, empty); err != nil {
			return 0, nil, nil, err
		}
	}
	if victimDead < 0 {
		return 0, nil, nil, nil
	}

	// Collect the victim's blocks until the move is full.
	var size int64
	err = t.itx.ScanPrefix(ctx, []byte(blockPrefix), func(key, value []byte) error {
		// Keep a block of the victim while the move has room.
		loc, err := unmarshalLocation(value)
		if err != nil || loc.segment != victim || size >= moveBytes {
			return err
		}
		keys, locs = append(keys, string(key)), append(locs, loc)
		size += loc.length
		return nil
	})
	return victim, keys, locs, err
}

// move copies the blocks at keys, read from locs in segment n, to the end of
// the log without the writer, then publishes the new location of each block
// still at its old one in one durable commit under the writer. A block
// changed meanwhile keeps its change, and its copy is dead space.
func (s *Store) move(ctx context.Context, n uint64, keys []string, locs []*location) error {
	// Read the payloads in one device call.
	reads := make([]device.Read, len(locs))
	for i, loc := range locs {
		reads[i] = device.Read{Name: segmentName(n), Offset: loc.offset, Data: make([]byte, loc.length)}
	}
	if err := s.dev.Read(ctx, reads); err != nil {
		return errors.Wrap(err, "read moved blocks")
	}

	// Write them at the end of the log.
	s.mtx.Lock()
	writes := make([]device.Write, len(reads))
	moved := make([]*location, len(reads))
	for i, r := range reads {
		moved[i] = s.place(int64(len(r.Data)))
		writes[i] = device.Write{Name: segmentName(moved[i].segment), Offset: moved[i].offset, Data: r.Data}
	}
	s.mtx.Unlock()
	if err := s.dev.Write(ctx, writes, false); err != nil {
		return err
	}

	// Take the writer and keep the copies of the blocks still in place.
	t, err := s.begin(ctx, true, nil)
	if err != nil {
		return err
	}
	defer t.Discard()
	for i, key := range keys {
		cur, err := getLocation(ctx, t.itx, []byte(key))
		if err != nil {
			return err
		}
		if cur == nil || *cur != *locs[i] {
			continue
		}
		s.mtx.Lock()
		if s.pending[key] == nil {
			t.own[key] = &pendingBlock{loc: moved[i]}
		}
		s.mtx.Unlock()
	}

	// Publish the new locations.
	return t.Commit(ctx)
}

// place returns the location of the next payload of length bytes, starting
// a new segment when the open one is full. The caller holds mtx.
func (s *Store) place(length int64) *location {
	if s.offset != 0 && s.offset+length > segmentSize {
		s.segment++
		s.offset = 0
	}
	loc := &location{segment: s.segment, offset: s.offset, length: length}
	s.offset += length
	return loc
}
