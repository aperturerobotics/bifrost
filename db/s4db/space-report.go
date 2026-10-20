package s4db

import (
	"context"
	"math/bits"
	"slices"
)

// SpaceReport details where a file's free space lies and what holds it, to
// explain a file longer than the data it holds.
type SpaceReport struct {
	// FreeRuns groups the free runs by length, shortest first.
	FreeRuns []FreeRunClass
	// LargestFreeBytes is the length of the longest free run.
	LargestFreeBytes int64
	// ValueEndBytes is the end of the highest page holding a live value.
	// Compaction can shorten the file toward it.
	ValueEndBytes int64
	// CompactGoalBytes is the end background compaction is moving the file
	// toward, or zero.
	CompactGoalBytes int64
	// PendingValueRuns counts the freed value runs a snapshot may still read.
	PendingValueRuns int
	// PendingPageRuns counts the freed index pages and log chunks a snapshot
	// may still read.
	PendingPageRuns int
	// ReservedBytes is the length of the runs this process's checkpoint is
	// writing into.
	ReservedBytes int64
	// UnpunchedBytes is the length of the runs freed since the last hole
	// punch.
	UnpunchedBytes int64
	// Slots lists the reader slots holding a pin.
	Slots []SlotPin
	// DiskBytes is the space the file system allocates to the file, or -1
	// when it cannot tell.
	DiskBytes int64
	// FreeDiskBytes is the part of DiskBytes inside free runs, which holes
	// were not punched for, or -1 when it cannot tell.
	FreeDiskBytes int64
}

// FreeRunClass counts the free runs of one length class.
type FreeRunClass struct {
	// MinBytes is the shortest run of the class. The longest is just under
	// twice it.
	MinBytes int64
	// Runs counts the runs.
	Runs int64
	// Bytes is the total length of the runs.
	Bytes int64
}

// SlotPin is the oldest state a reader slot's snapshots read.
type SlotPin struct {
	// Slot is the slot index.
	Slot int
	// Seq is the oldest commit the snapshots read.
	Seq uint64
	// Ckpt is the oldest checkpoint the snapshots read.
	Ckpt uint64
	// Live reports whether an open handle holds the slot. A pin in a slot
	// no handle holds keeps nothing allocated.
	Live bool
	// Own reports whether the slot is this handle's.
	Own bool
}

// allocator is a storage that can report the bytes it allocates.
type allocator interface {
	// allocated returns the bytes in [lo, hi) the file system holds data
	// for, reporting false when it cannot tell.
	allocated(lo, hi int64) (int64, bool)
}

// SpaceReport reports the layout of the file's free space as of the latest
// commit. It holds the writer lock only to copy the space.
func (db *DB) SpaceReport(ctx context.Context) (SpaceReport, error) {
	// Take the writer lock so the space is current and stable.
	if err := db.w.lock(ctx); err != nil {
		return SpaceReport{}, err
	}
	sp := db.w.sp

	// Copy the free runs and the counts of what holds space back.
	var frees []run
	sp.free.Scan(func(start, n uint64) bool {
		frees = append(frees, run{start: start, n: n})
		return true
	})
	r := SpaceReport{
		CompactGoalBytes: pageOff(sp.goal),
		PendingValueRuns: len(sp.values),
		PendingPageRuns:  len(sp.pages),
		UnpunchedBytes:   pageOff(sp.releasedPages),
	}
	for _, rr := range sp.reserved {
		r.ReservedBytes += pageOff(rr.n)
	}

	// Find the end of the highest live value page, then release the lock.
	for i, v := range slices.Backward(sp.live) {
		if v != 0 {
			r.ValueEndBytes = pageOff(uint64(i) + 1) // #nosec G115 -- slice indexes are not negative.
			break
		}
	}
	end := pageOff(sp.end)
	db.w.unlock(nil)

	// Group the free runs by the power of two below their length.
	var classes [64]FreeRunClass
	for _, fr := range frees {
		k := bits.Len64(fr.n) - 1
		classes[k].Runs++
		classes[k].Bytes += pageOff(fr.n)
		r.LargestFreeBytes = max(r.LargestFreeBytes, pageOff(fr.n))
	}
	for k, c := range classes {
		if c.Runs != 0 {
			c.MinBytes = pageOff(1 << k)
			r.FreeRuns = append(r.FreeRuns, c)
		}
	}

	// Read the slots and whether a handle holds each.
	b := make([]byte, pageSize)
	if _, err := db.s.ReadAt(b, slotPage*pageSize); err != nil {
		return SpaceReport{}, err
	}
	for i := range slots {
		p, ok := decodePin(b[i*slotSize:])
		if !ok {
			continue
		}
		own := i == db.slot
		live := own
		if !own {
			held, err := db.s.held(lockSlot + int64(i))
			if err != nil {
				return SpaceReport{}, err
			}
			live = held
		}
		r.Slots = append(r.Slots, SlotPin{Slot: i, Seq: p.seq, Ckpt: p.ckpt, Live: live, Own: own})
	}

	// Ask the file system what it allocates, in the whole file and inside
	// the free runs.
	r.DiskBytes, r.FreeDiskBytes = -1, -1
	if a, ok := db.s.(allocator); ok {
		if n, ok := a.allocated(0, end); ok {
			r.DiskBytes = n
			r.FreeDiskBytes = 0
			for _, fr := range frees {
				n, ok := a.allocated(pageOff(fr.start), pageOff(fr.start+fr.n))
				if !ok {
					r.FreeDiskBytes = -1
					break
				}
				r.FreeDiskBytes += n
			}
		}
	}
	return r, nil
}
