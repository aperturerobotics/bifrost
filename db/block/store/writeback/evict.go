package block_store_writeback

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/net/hash"
)

// evictPrefix prefixes the eviction queue entry of each uploaded block, ordered
// by the pass that queued it: evictable/<16 hex digits of epoch>/<hash>. The
// value is the block size.
const evictPrefix = "evictable/"

// epochKey holds the current eviction epoch, which each pass advances.
const epochKey = "epoch"

// evictableBytesKey holds the summed size of the eviction queue entries.
const evictableBytesKey = "evictable-bytes"

// evictBatchBlocks bounds the entries one eviction step reads and commits.
const evictBatchBlocks = 256

// EvictResult reports what an eviction pass did.
type EvictResult struct {
	// Evicted is the number of local values dropped.
	Evicted int
	// EvictedBytes is the recorded size of the dropped values.
	EvictedBytes int64
	// Requeued is the number of entries the caller kept, queued at the new
	// epoch.
	Requeued int
	// Refused is the number of blocks kept because the remote store may not
	// hold them: a marker awaits upload, or the remote does not report them.
	Refused int
}

// evictEntry is a decoded eviction queue entry.
type evictEntry struct {
	// key is the entry's key, owned by the entry.
	key []byte
	// epoch is the pass that queued the entry.
	epoch uint64
	// hash is the block hash.
	hash *hash.Hash
	// size is the block size in bytes.
	size int64
}

// evictKey returns the eviction queue key of a block queued at epoch.
func evictKey(epoch uint64, h *hash.Hash) []byte {
	// Encode the epoch as fixed-width hex, so keys sort by epoch.
	var stamp [8]byte
	binary.BigEndian.PutUint64(stamp[:], epoch)
	key := []byte(evictPrefix)
	key = hex.AppendEncode(key, stamp[:])
	key = append(key, '/')
	return append(key, h.MarshalString()...)
}

// parseEvictKey decodes an eviction queue key into an entry without its size.
func parseEvictKey(key []byte) (evictEntry, error) {
	// Split the key into its epoch and hash.
	rest := key[len(evictPrefix):]
	if len(rest) <= queueTimeLen || rest[queueTimeLen] != '/' {
		return evictEntry{}, errors.Errorf("malformed eviction queue key %q", key)
	}

	// Decode the epoch and the block hash.
	var stamp [8]byte
	if _, err := hex.Decode(stamp[:], rest[:queueTimeLen]); err != nil {
		return evictEntry{}, errors.Wrap(err, "parse eviction queue epoch")
	}
	h := &hash.Hash{}
	if err := h.ParseFromB58(string(rest[queueTimeLen+1:])); err != nil {
		return evictEntry{}, errors.Wrap(err, "parse eviction queue hash")
	}
	return evictEntry{key: bytes.Clone(key), epoch: binary.BigEndian.Uint64(stamp[:]), hash: h}, nil
}

// scanEvictable calls cb with each eviction queue entry, oldest epoch first,
// until cb returns an error.
func scanEvictable(ctx context.Context, tx kvtx.Tx, cb func(evictEntry) error) error {
	it := tx.Iterate(ctx, []byte(evictPrefix), true, false)
	defer it.Close()
	for it.Next() {
		entry, err := parseEvictKey(it.Key())
		if err != nil {
			return err
		}
		value, err := it.Value()
		if err != nil {
			return err
		}
		entry.size, _ = strconv.ParseInt(string(value), 10, 64)
		if err := cb(entry); err != nil {
			return err
		}
	}
	return it.Err()
}

// queueEvict sets a block's eviction queue entry at epoch, returning the bytes
// it added to the queue: zero when the entry was already there.
func queueEvict(ctx context.Context, tx kvtx.Tx, epoch uint64, h *hash.Hash, size int64) (int64, error) {
	key := evictKey(epoch, h)
	found, err := tx.Exists(ctx, key)
	if err != nil || found {
		return 0, err
	}
	return size, tx.Set(ctx, key, []byte(strconv.FormatInt(size, 10)))
}

// readEvictableBytes reads the summed size of the eviction queue entries.
func readEvictableBytes(ctx context.Context, tx kvtx.Tx) (int64, error) {
	value, found, err := tx.Get(ctx, []byte(evictableBytesKey))
	if err != nil || !found {
		return 0, err
	}
	return strconv.ParseInt(string(value), 10, 64)
}

// addEvictableBytes adds delta to the summed size of the eviction queue
// entries, in the transaction that changed them.
func addEvictableBytes(ctx context.Context, tx kvtx.Tx, delta int64) error {
	if delta == 0 {
		return nil
	}
	total, err := readEvictableBytes(ctx, tx)
	if err != nil {
		return err
	}
	return tx.Set(ctx, []byte(evictableBytesKey), []byte(strconv.FormatInt(total+delta, 10)))
}

// EvictBytes returns how many uploaded bytes a pass drops to bring the
// resident bytes from past the high mark down to the low mark, or zero.
func (s Status) EvictBytes() int64 {
	resident := s.ResidentBytes()
	if s.Policy.ResidentHigh <= 0 || resident <= s.Policy.ResidentHigh {
		return 0
	}
	return min(resident-s.Policy.ResidentLow, s.EvictableBytes)
}

// addEvictStatus projects committed eviction queue changes into the status:
// deltas commute, so concurrent transactions project in any order. A change
// of nothing wakes no watcher, so a budget watcher whose pass changed nothing
// waits for the next write.
func (s *Store) addEvictStatus(evictable, evicted, refills int64) {
	if evictable == 0 && evicted == 0 && refills == 0 {
		return
	}
	s.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		s.status.EvictableBytes += evictable
		s.status.EvictedBytes += evicted
		s.status.Refills += refills
		broadcast()
	})
}

// readEpoch reads the current eviction epoch, zero before the first pass.
func readEpoch(ctx context.Context, tx kvtx.Tx) (uint64, error) {
	value, found, err := tx.Get(ctx, []byte(epochKey))
	if err != nil || !found {
		return 0, err
	}
	return strconv.ParseUint(string(value), 10, 64)
}

// Refill writes a block read back from the remote store into the local
// values and queues it for eviction at the current epoch. It changes no
// ownership, so call it only for a block the reference graph still tracks.
func (s *Store) Refill(ctx context.Context, ref *block.BlockRef, stored *block.StoredBlock) error {
	// Write the value below the ownership layer.
	if _, _, err := s.values.PutBlock(ctx, stored.Data, stored.PutOpts(ref)); err != nil {
		return errors.Wrap(err, "refill block")
	}

	// Queue the refilled block for eviction.
	var added int64
	err := kvtx.RunTransaction(ctx, true, s.writeTx, func(ctx context.Context, tx kvtx.Tx) error {
		// Queue the entry at the current epoch and count its bytes.
		epoch, err := readEpoch(ctx, tx)
		if err != nil {
			return err
		}
		added, err = queueEvict(ctx, tx, epoch, ref.GetHash(), int64(len(stored.Data)))
		if err != nil {
			return err
		}
		return addEvictableBytes(ctx, tx, added)
	})
	if err != nil {
		return errors.Wrap(err, "queue refilled block")
	}
	s.addEvictStatus(added, 0, 1)
	return nil
}

// Evict runs one eviction pass, one at a time: it advances the epoch, then walks the queue
// oldest epoch first and drops local values until it has dropped maxBytes or
// reaches the new epoch. A block keep reports is queued at the new epoch
// instead. A nil keep keeps nothing.
//
// A block leaves the local values only while no marker awaits its upload and
// remote reports holding it, so the remote store always holds what was
// dropped. An entry whose value is already gone, such as a block the garbage
// collector swept, is deleted. Values are dropped before their entries, so an
// interrupted pass leaves only entries the next pass deletes. The pass takes
// no lock a write waits on, since asking remote can wait on the network.
//
// The volume shares values by hash across its block stores. Eviction relies on
// each placed store sealing its blocks with its own transform keys, so no other
// store holds the same hash.
func (s *Store) Evict(
	ctx context.Context,
	remote block.StoreOps,
	maxBytes int64,
	keep func(*hash.Hash) bool,
) (EvictResult, error) {
	// Serialize the pass with other passes.
	var result EvictResult
	release, err := s.evictMtx.Lock(ctx)
	if err != nil {
		return result, err
	}
	defer release()

	// Advance the epoch, so entries this pass queues sort after the walk.
	var epoch uint64
	err = kvtx.RunTransaction(ctx, true, s.writeTx, func(ctx context.Context, tx kvtx.Tx) error {
		prev, err := readEpoch(ctx, tx)
		if err != nil {
			return err
		}
		epoch = prev + 1
		return tx.Set(ctx, []byte(epochKey), []byte(strconv.FormatUint(epoch, 10)))
	})
	if err != nil {
		return result, errors.Wrap(err, "advance eviction epoch")
	}

	// Evict in batches until the byte target or the new epoch.
	for result.EvictedBytes < maxBytes {
		done, err := s.evictBatch(ctx, remote, epoch, maxBytes, keep, &result)
		if err != nil || done {
			return result, err
		}
	}
	return result, nil
}

// evictBatch reads one batch of entries older than epoch, drops the values it
// may, and commits the entry changes. Returns true when the walk reached the
// end of the older entries.
func (s *Store) evictBatch(
	ctx context.Context,
	remote block.StoreOps,
	epoch uint64,
	maxBytes int64,
	keep func(*hash.Hash) bool,
	result *EvictResult,
) (bool, error) {
	// Read the batch with the markers of its blocks.
	var batch []evictEntry
	var marked []bool
	var done bool
	err := kvtx.RunTransaction(ctx, false, s.readTx, func(ctx context.Context, tx kvtx.Tx) error {
		// Walk the older entries up to the batch and byte limits; the walk is
		// done when it runs out of them.
		batch, marked, done = batch[:0], marked[:0], true
		var batchBytes int64
		err := scanEvictable(ctx, tx, func(entry evictEntry) error {
			// Stop at the new epoch or a full batch.
			if entry.epoch >= epoch || len(batch) == evictBatchBlocks || result.EvictedBytes+batchBytes >= maxBytes {
				done = entry.epoch >= epoch
				return errScanDone
			}

			// Take the entry with whether a marker awaits its upload.
			found, err := tx.Exists(ctx, markedKey(entry.hash))
			if err != nil {
				return err
			}
			batch, marked = append(batch, entry), append(marked, found)
			batchBytes += entry.size
			return nil
		})
		if errors.Is(err, errScanDone) {
			return nil
		}
		return err
	})
	if err != nil {
		return false, errors.Wrap(err, "scan eviction queue")
	}
	if len(batch) == 0 {
		return true, nil
	}

	// Ask the local values and the remote which blocks they hold.
	refs := make([]*block.BlockRef, len(batch))
	for i, entry := range batch {
		refs[i] = block.NewBlockRef(entry.hash)
	}
	held, err := s.values.GetBlockExistsBatch(ctx, refs)
	if err != nil {
		return false, errors.Wrap(err, "check local values")
	}
	stored, err := remote.GetBlockExistsBatch(ctx, refs)
	if err != nil {
		return false, errors.Wrap(err, "check remote blocks")
	}

	// Drop the values the remote holds, and decide each entry.
	var requeue []evictEntry
	var evictedBytes int64
	for i, entry := range batch {
		switch {
		case !held[i]:
		case marked[i] || !stored[i]:
			result.Refused++
		case keep != nil && keep(entry.hash):
			requeue = append(requeue, entry)
		default:
			if err := s.values.RmBlock(ctx, refs[i]); err != nil {
				return false, errors.Wrap(err, "evict block")
			}
			result.Evicted++
			evictedBytes += entry.size
		}
	}

	// Delete the walked entries and queue the kept ones at the new epoch. A
	// refused block's next upload queues it again.
	var delta int64
	err = kvtx.RunTransaction(ctx, true, s.writeTx, func(ctx context.Context, tx kvtx.Tx) error {
		delta = 0
		for _, entry := range batch {
			if err := tx.Delete(ctx, entry.key); err != nil {
				return err
			}
			delta -= entry.size
		}
		for _, entry := range requeue {
			added, err := queueEvict(ctx, tx, epoch, entry.hash, entry.size)
			if err != nil {
				return err
			}
			delta += added
		}
		return addEvictableBytes(ctx, tx, delta)
	})
	if err != nil {
		return false, errors.Wrap(err, "update eviction queue")
	}

	// Report the batch.
	s.addEvictStatus(delta, evictedBytes, 0)
	result.EvictedBytes += evictedBytes
	result.Requeued += len(requeue)
	return done, nil
}
