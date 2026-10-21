package s4db

import (
	"bytes"
	"context"
	"math"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/aperturerobotics/util/csync"
	"github.com/pkg/errors"
)

// compactBatch bounds the value bytes Compact holds in memory and commits
// at once.
const compactBatch = 64 << 20

// writer appends commits, checkpoints, and moves values for one handle. Its
// fields belong to the holder of the writer lock.
type writer struct {
	// db is the database.
	db *DB
	// mtx excludes this handle's other writers.
	mtx csync.Mutex
	// unlockMtx releases mtx while the writer lock is held.
	unlockMtx func()
	// sp is the space state, current through spSeq and spGen. Replacing it
	// also takes spMtx.
	sp *space
	// spMtx excludes replacing sp while a background checkpoint writes a
	// page it reserved from sp, which a single-process handle does without
	// the writer lock.
	spMtx sync.Mutex
	// spSeq is the commit sp reflects.
	spSeq uint64
	// spGen is the checkpoint generation sp reflects.
	spGen uint64
	// fileEnd is the file length in pages.
	fileEnd uint64
	// relocKey is where the next compaction pass resumes.
	relocKey []byte
	// ckptErr is the error of the last background checkpoint, reported by
	// the next commit.
	ckptErr error

	// held is set while this handle holds the writer lock. It is safe
	// without the writer lock.
	held atomic.Bool
	// waiting counts this handle's writers waiting for the writer lock. It
	// is safe without the writer lock.
	waiting atomic.Int32
	// building is set while a background checkpoint runs. It and builds
	// are safe without the writer lock.
	building atomic.Bool
	// builds counts running background checkpoints, so Close can wait.
	builds sync.WaitGroup
	// buildLocked is set while this handle's background checkpoint holds
	// a shared file's build lock. It is set under the writer lock.
	buildLocked atomic.Bool
	// closing is set by Close, so a running background checkpoint drops
	// its tree instead of finishing it.
	closing atomic.Bool
	// changed lists the commit records published since the state a
	// background checkpoint builds.
	changed []*record
}

// lock takes the writer lock and brings the writer up to date with records
// other processes appended. Waiting for another process cannot be
// canceled; ctx bounds the wait for this handle's writers.
func (w *writer) lock(ctx context.Context) error {
	// Exclude this handle's writers, then other processes, counted as
	// waiting until both are held.
	w.waiting.Add(1)
	release, err := w.mtx.Lock(ctx)
	if err != nil {
		w.waiting.Add(-1)
		return err
	}
	_, err = w.db.lockFile(true)
	w.waiting.Add(-1)
	if err != nil {
		release()
		return err
	}
	w.unlockMtx = release
	w.held.Store(true)

	// Apply what other processes wrote.
	if err := w.catchUp(ctx); err != nil {
		w.unlock(err)
		return err
	}
	return nil
}

// tryLock takes the writer lock when no writer holds it, without catching
// up.
func (w *writer) tryLock() (bool, error) {
	// Exclude this handle's writers, then other processes, without
	// waiting.
	release, ok := w.mtx.TryLock()
	if !ok {
		return false, nil
	}
	ok, err := w.db.lockFile(false)
	if err != nil || !ok {
		release()
		return false, err
	}
	w.unlockMtx = release
	w.held.Store(true)
	return true, nil
}

// unlock releases the writer lock and tells other processes the file
// changed. err is the result of the work done under the lock: a failed
// step may have changed the space for writes the log does not hold, so
// the space is dropped and the next lock rebuilds it from the log.
func (w *writer) unlock(err error) {
	// Drop a space a failed step may have left ahead of the log.
	if err != nil {
		w.setSpace(nil)
	}

	// Release other processes and wake their watchers, then this handle's
	// writers.
	w.held.Store(false)
	_ = w.db.unlockFile()
	w.db.watch.notify()
	release := w.unlockMtx
	w.unlockMtx = nil
	release()
}

// setSpace replaces the space once no background checkpoint is writing a
// page reserved from it. The caller holds the writer lock.
func (w *writer) setSpace(sp *space) {
	w.spMtx.Lock()
	w.sp = sp
	w.spMtx.Unlock()
}

// create writes an empty database into an empty file. The caller holds the
// file's writer lock.
func (w *writer) create() error {
	// Leave an existing file alone.
	s := w.db.s
	if size, err := s.size(); err != nil || size != 0 {
		return err
	}

	// Write the header and an empty slot table.
	if _, err := s.WriteAt(newHeader(w.db.opts.InlineMax).encode(), headerPage*pageSize); err != nil {
		return err
	}
	if _, err := s.WriteAt(make([]byte, pageSize), slotPage*pageSize); err != nil {
		return err
	}

	// Allocate the first log chunk and save the space.
	sp := newSpace()
	chunk := sp.alloc(extentPages)
	sp.chunks = []run{chunk}
	ref, err := w.writeSpace(sp)
	if err != nil {
		return err
	}

	// Point the first superblock at the empty log and flush.
	sb := superblock{gen: 1, space: ref, logPos: chunk.start * pageSize, logEnd: chunk.end() * pageSize}
	if err := s.truncate(pageOff(sp.end)); err != nil {
		return err
	}
	if _, err := s.WriteAt(sb.encode(), sb.page()*pageSize); err != nil {
		return err
	}
	return s.flushDurable()
}

// catchUp applies new records and rebuilds the space when another process
// wrote since this handle last did. The caller holds the writer lock.
func (w *writer) catchUp(ctx context.Context) error {
	// Apply the new records and keep the space when it is still current.
	db := w.db
	if err := db.Refresh(ctx); err != nil {
		return err
	}
	st := db.cur.Load()
	if w.sp != nil && w.spSeq == st.seq && w.spGen == st.gen {
		return nil
	}
	if err := w.replaySpace(st); err != nil {
		return err
	}

	// Flush whatever the previous writer left unflushed, so this writer's
	// durability marks hold.
	return db.flush.syncThrough(ctx, st.seq, st.ckpt)
}

// replaySpace loads the space of st's checkpoint and replays the records
// after it.
func (w *writer) replaySpace(st *state) error {
	// Load the checkpoint's space.
	s := w.db.s
	ref := st.sb.space
	b := make([]byte, ref.n)
	if _, err := s.ReadAt(b, fileOff(ref.off)); err != nil {
		return err
	}
	if checksum(b) != ref.crc {
		return errors.Wrap(ErrCorrupt, "space checksum mismatch")
	}
	sp, err := decodeSpace(b)
	if err != nil {
		return err
	}

	// Replay the space changes of the records after it, indexing the
	// pending runs so each claim finds the runs it overlaps. The log resumes
	// after the tree's checkpoint, which may precede the space's.
	sp.indexPending()
	lr := newLogReader(s, st.sb.logPos, st.sb.logEnd, st.sb.seq, st.sb.logCrc)
	for lr.seq <= st.seq {
		r, ok := lr.next()
		if !ok {
			return errors.Wrapf(ErrCorrupt, "log ends before record %d", st.seq)
		}
		if r.seq <= st.sb.spaceSeq {
			continue
		}
		if r.kind == kindLink {
			sp.claim(r.next)
			sp.chunks = append(sp.chunks, r.next)
			continue
		}

		// Release first: the writer placed this record's values after
		// every release its marks cover.
		sp.release(r.released.seq, r.released.ckpt)
		for _, o := range r.ops {
			if o.val.isRef {
				sp.useValue(o.val.ref.off, int(o.val.ref.n))
			}
		}
		for _, fr := range r.frees {
			sp.freeValue(fr.off, int(fr.n), r.seq)
		}
	}

	// Adopt the space. Its writer punched what it released, except a last
	// batch a crash cut short, whose pages stay allocated until reused.
	sp.pendingAt = nil
	sp.drain()
	size, err := s.size()
	if err != nil {
		return err
	}
	w.setSpace(sp)
	w.spSeq, w.spGen = st.seq, st.gen
	w.fileEnd = uint64(size) / pageSize // #nosec G115 -- file sizes are not negative.
	return nil
}

// commit appends a commit record for changes on base, writes its values,
// and publishes the new state. The caller holds the writer lock.
func (w *writer) commit(ctx context.Context, base *state, changes []tentry, ordered bool) error {
	// Load the index paths of the changes the overlay does not answer
	// together, so a cold commit reads one round per tree level instead of
	// one path per change while it holds the writer lock.
	db, sp := w.db, w.sp
	var cold [][]byte
	for _, c := range changes {
		if !base.filter.has(c.key) {
			cold = append(cold, c.key)
		}
	}
	if err := db.p.prefetch(ctx, base.root, cold); err != nil {
		return err
	}

	// Resolve each change against the base state: frees, count, and value
	// placement.
	r := &record{kind: kindCommit}
	var writes []valueWrite
	for _, c := range changes {
		old, found, err := db.lookup(base, c.key)
		if err != nil {
			return err
		}
		if c.del && !found {
			continue
		}
		if found && old.isRef {
			r.frees = append(r.frees, old.ref)
		}
		switch {
		case c.del:
			r.delta--
			r.ops = append(r.ops, op{key: c.key, del: true})
			continue
		case !found:
			r.delta++
		}
		v := value{inline: c.data}
		if len(c.data) > int(db.hdr.inlineMax) {
			v = value{isRef: true, ref: extentRef{n: uint32(len(c.data)), crc: checksum(c.data)}} // #nosec G115 -- Set bounds values by MaxValueSize.
			writes = append(writes, valueWrite{op: len(r.ops), data: c.data})
		}
		r.ops = append(r.ops, op{key: c.key, val: v})
	}
	if len(r.ops) == 0 {
		return nil
	}

	// Refuse a record its length field cannot hold.
	size := r.bound() + linkSize
	if int64(size) > maxRecord {
		return ErrTxTooLarge
	}

	// Continue the log in a new chunk when the record may not fit.
	bound := uint64(size) // #nosec G115 -- sizes are not negative.
	seq := base.seq + 1
	pos, end, prev := base.pos, base.end, base.crc
	var link []byte
	linkPos := pos
	if pos+bound > end {
		chunk := sp.alloc(max(extentPages, pagesFor(size)))
		sp.chunks = append(sp.chunks, chunk)
		l := newLink(seq, chunk)
		link = l.encode(prev)
		seq, prev = seq+1, l.crc
		pos, end = chunk.start*pageSize, chunk.end()*pageSize
	}
	r.seq = seq
	w.db.ownSeq.Store(seq)

	// Place values, then free replaced ones and release what no snapshot
	// reads.
	for _, wr := range writes {
		ref := &r.ops[wr.op].val.ref
		ref.off = sp.placeValue(len(wr.data), seq)
	}
	for _, fr := range r.frees {
		sp.freeValue(fr.off, int(fr.n), seq)
	}
	if err := w.releaseSpace(ctx, base, r); err != nil {
		return err
	}

	// Write the values, then the record. A tail of this handle that began
	// before the writer lock and reads the record first publishes the same
	// state.
	if err := w.writeValues(r, writes); err != nil {
		return err
	}
	r.durable, _, _ = db.flush.marks()
	rec := r.encode(prev)
	if err := w.writeRecord(link, linkPos, rec, pos); err != nil {
		return err
	}

	// Order an ordered commit's writes before the next commit's. A failure
	// leaves the state unpublished; erasing the header at linkPos ends the
	// log before the record, so a later issue of its writes cannot make it
	// readable.
	if ordered {
		if err := db.s.flushOrdered(); err != nil {
			_, _ = db.s.WriteAt(make([]byte, recordHeader), fileOff(linkPos))
			return err
		}
	}

	// Publish the new state.
	next := db.derive(base)
	next.apply(r)
	next.pos, next.end = pos+uint64(len(rec)), end
	if err := db.publish(ctx, next); err != nil {
		return err
	}
	if w.building.Load() {
		w.changed = append(w.changed, r)
	}

	// Arm the deadline flush of an ordered commit and release space once a
	// batch has built up. A durable commit flushes after the writer lock is
	// released.
	if ordered {
		db.flush.ordered()
	}
	release := w.truncate
	if w.sp.releasedPages >= punchBatchPages {
		release = w.punch
	}
	if err := release(); err != nil {
		return err
	}

	// The space now reflects the record.
	w.spSeq = seq
	return nil
}

// writeRecord writes a commit record at pos, after the link that leads to
// its chunk when link is set.
func (w *writer) writeRecord(link []byte, linkPos uint64, rec []byte, pos uint64) error {
	if link != nil {
		if _, err := w.db.s.WriteAt(link, fileOff(linkPos)); err != nil {
			return err
		}
	}
	_, err := w.db.s.WriteAt(rec, fileOff(pos))
	return err
}

// writeValues writes placed values, joining adjacent ones into one write.
func (w *writer) writeValues(r *record, writes []valueWrite) error {
	// Collect adjacent values in buf, starting at at.
	var buf []byte
	var at uint64
	flush := func() error {
		if len(buf) == 0 {
			return nil
		}
		_, err := w.db.s.WriteAt(buf, fileOff(at))
		buf = buf[:0]
		return err
	}

	// Write each run when the next value is not adjacent.
	for _, wr := range writes {
		off := r.ops[wr.op].val.ref.off
		if len(buf) != 0 && at+uint64(len(buf)) != off {
			if err := flush(); err != nil {
				return err
			}
		}
		if len(buf) == 0 {
			at = off
		}
		buf = append(buf, wr.data...)
	}
	return flush()
}

// releaseSpace returns freed runs no snapshot reads to free space and
// records the space's release marks in r. The marks only rise, so replaying
// any record repeats every release the writer made before it.
func (w *writer) releaseSpace(ctx context.Context, base *state, r *record) error {
	// Bound the release by durability and this handle's snapshots, and skip
	// reading the slots when nothing pending falls under that bound.
	sp := w.sp
	defer func() { r.released = sp.marks }()
	_, logged, ckptDurable := w.db.flush.marks()
	local := w.db.localPin()
	seq := min(local.seq, logged)
	ckpt := min(local.ckpt, base.ckpt, ckptDurable)
	if !sp.due(seq, ckpt) {
		return nil
	}

	// Lower the bound by every other process's snapshots and release value
	// runs and pages below it.
	p, err := w.db.minPin()
	if err != nil {
		return err
	}
	sp.release(min(p.seq, seq), min(p.ckpt, ckpt))
	return nil
}

// punchBatchPages is the released page count at which a commit punches.
// Each punch is a file system call made under the writer lock, and
// lowest-fit allocation soon takes freed runs again, so commits leave smaller
// batches to the next checkpoint.
const punchBatchPages = 8 << 20 / pageSize

// punch deallocates released runs and shrinks the file to the space end.
func (w *writer) punch() error {
	// Deallocate released runs below the end.
	db, sp := w.db, w.sp
	for _, r := range sp.drain() {
		if r.start >= sp.end {
			continue
		}
		r.n = min(r.n, sp.end-r.start)
		db.p.cache.drop(r)
		if err := db.s.punch(pageOff(r.start), pageOff(r.n)); err != nil {
			return err
		}
	}
	return w.truncate()
}

// truncate shrinks the file to the space end.
func (w *writer) truncate() error {
	db, sp := w.db, w.sp
	if sp.end < w.fileEnd {
		db.p.cache.drop(run{start: sp.end, n: w.fileEnd - sp.end})
		if err := db.s.truncate(pageOff(sp.end)); err != nil {
			return err
		}
	}
	w.fileEnd = sp.end
	return nil
}

// checkpoint writes the overlay into the tree and saves a superblock. The
// caller holds the writer lock.
func (w *writer) checkpoint(ctx context.Context) error {
	// Skip when no commit follows the last checkpoint.
	st := w.db.cur.Load()
	if st.seq == st.ckpt {
		return nil
	}

	// Build and write the tree, reserving its pages as it goes, and save
	// it.
	t := w.newTree(st)
	t.b.out.reserve = func(n uint64) (run, error) {
		return w.reserve(t, n)
	}
	if err := t.build(); err != nil {
		return err
	}
	return w.saveTree(ctx, t)
}

// startCheckpoint builds the tree of the published state in the
// background, then saves it under the writer lock, so commits continue
// while the tree builds. At most one build runs. The caller holds the
// writer lock.
func (w *writer) startCheckpoint() {
	// Claim the build and keep the state's pages readable until the tree
	// is saved.
	if !w.building.CompareAndSwap(false, true) {
		return
	}

	// In a shared file, build only while no other process does. Its tree
	// covers this handle's records too.
	if !w.db.opts.SingleProcess {
		ok, err := w.db.s.lock(lockBuild, false)
		if err != nil || !ok {
			w.building.Store(false)
			return
		}
		w.buildLocked.Store(true)
	}
	w.changed = nil
	st, stripe := w.db.acquire()

	// Skip a state no commit follows since its checkpoint, as checkpoint
	// does. Its tree would free pages under the checkpoint that snapshots
	// of the state pin, so a release would hand them out while they read.
	if st.seq == st.ckpt {
		w.db.release(st, stripe)
		w.unlockBuild()
		w.endBuild()
		return
	}

	// Build, then save and release under the writer lock.
	w.builds.Go(func() {
		w.runCheckpoint(st, stripe)
		w.endBuild()
	})
}

// endBuild ends the claim on the build and wakes the writers waiting for it.
func (w *writer) endBuild() {
	w.building.Store(false)
	w.db.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		broadcast()
	})
}

// waitBuild waits while a background checkpoint runs and the published
// overlay exceeds twice CheckpointMax, or the log after the checkpoint four
// times, so a writer that outpaces checkpoints holds bounded memory and
// replay. The caller does not hold the writer lock, which the build needs.
func (w *writer) waitBuild(ctx context.Context) error {
	// Return at once while the state is within bounds.
	db := w.db
	limit := 2 * db.opts.CheckpointMax
	within := func() bool {
		st := db.cur.Load()
		return st.overlayBytes <= limit && st.logBytes <= 2*limit
	}
	if within() {
		return nil
	}

	// Wait for the build to end or the state to come back within bounds.
	return db.bcast.Wait(ctx, func(_ func(), _ func() <-chan struct{}) (bool, error) {
		return !w.building.Load() || within(), nil
	})
}

// runCheckpoint builds the tree of st, which the snapshot in stripe keeps,
// and writes its pages, holding the writer lock only to reserve pages, to
// write them to a shared file, and to save the tree. The next commit reports a failure.
func (w *writer) runCheckpoint(st *state, stripe int) {
	// Reserve the tree's pages under the writer lock as it builds.
	ctx := context.Background()
	t := w.newTree(st)
	out := t.b.out
	out.reserve = func(n uint64) (run, error) {
		if err := w.lock(ctx); err != nil {
			return run{}, err
		}
		r, err := w.reserve(t, n)
		w.unlock(nil)
		return r, err
	}

	// Write each batch of pages while the space still holds their
	// reservations. A failed commit reloads the space, and the space saved
	// in the file does not hold the reservations, so the reloaded space
	// may hand them out. Another process may also hand them out, so a
	// shared file is written under the writer lock, whose catch-up reloads
	// the space when another process wrote, and the build lock keeps other
	// processes' checkpoints and punches off them. A single-process
	// handle holds the file's writer lock for its lifetime and writes
	// without waiting for this handle's writers.
	write := out.write
	out.write = func(buf []byte, page uint64) error {
		// Exclude other processes' writers from a shared file.
		if !w.db.opts.SingleProcess {
			if err := w.lock(ctx); err != nil {
				return err
			}
			defer w.unlock(nil)
		}

		// Write while the space holds the reservations.
		w.spMtx.Lock()
		defer w.spMtx.Unlock()
		if t.sp != w.sp || w.closing.Load() {
			return errDropped
		}
		return write(buf, page)
	}

	// Build the tree, then flush its pages without the writer lock, so the
	// barrier under the lock flushes little.
	err := t.build()
	if err == nil {
		_, err = w.db.s.flushBarrier()
	}
	if err == nil {
		err = w.carryOver(ctx, t)
	}

	// Take the writer lock to save the tree.
	if lerr := w.lock(ctx); lerr != nil {
		w.db.release(st, stripe)
		w.unlockBuild()
		return
	}

	// Save the tree, then end the snapshot so the pages it replaced can be
	// released. A tree dropped while it built is no failure.
	if errors.Is(err, errDropped) {
		t.evict()
		err = nil
	} else if err == nil {
		err = w.saveTree(ctx, t)
	}
	w.db.release(st, stripe)
	if err == nil {
		err = w.releaseNow(ctx)
	}
	w.ckptErr = err
	w.unlockBuild()
	w.unlock(err)
}

// unlockBuild releases the build lock this handle's background checkpoint
// holds in a shared file.
func (w *writer) unlockBuild() {
	if w.buildLocked.Swap(false) {
		_ = w.db.s.unlock(lockBuild)
	}
}

// tree is a checkpoint tree built from a published state.
type tree struct {
	// st is the state whose overlay the tree includes.
	st *state
	// b writes the new pages and lists the pages they replace.
	b *builder
	// top is the new root, nil when the tree is empty.
	top *child
	// sp is the space holding the reserved pages, nil before the first
	// reservation.
	sp *space
	// carried holds the overlay entries after st through record
	// carriedSeq; nil before carryOver.
	carried *carried
	// carriedSeq is the last record carried holds.
	carriedSeq uint64
}

// errDropped reports a tree abandoned while it built: either its space was
// reloaded, which does not hold its reservations, so another process may
// have reused them, or the database is closing. The space saved in the file
// never holds the reservations, so the next open drops them either way.
var errDropped = errors.New("checkpoint space reloaded")

// newTree returns the tree of st, writing its pages through the database
// file and cache. The caller sets the writer's reserve function.
func (w *writer) newTree(st *state) *tree {
	out := &pageWriter{cache: w.db.p.cache, write: func(buf []byte, page uint64) error {
		_, err := w.db.s.WriteAt(buf, pageOff(page))
		return err
	}}
	return &tree{st: st, b: &builder{p: w.db.p, out: out}}
}

// build applies the overlay of t's state to its tree and writes the new
// pages. It reads only the state and the pages its tree holds, so it needs
// no writer lock while a snapshot keeps the state.
func (t *tree) build() error {
	// Collect the overlay in key order and apply it.
	st := t.st
	changes := make([]change, 0, st.overlay.Len())
	st.overlay.Scan(func(it overlayItem) bool {
		e := it.e
		changes = append(changes, change{key: e.key, del: e.del, val: e.val})
		return true
	})
	top, err := t.b.build(st.root, changes)
	if err != nil {
		return err
	}
	t.top = top
	return t.b.place(top)
}

// evict drops t's pages from the cache when t is dropped: no reader has
// seen them, and the space hands them out again.
func (t *tree) evict() {
	for _, r := range t.b.out.runs {
		t.b.out.cache.drop(r)
	}
}

// reserve takes up to n pages from the space for t's new pages. The caller
// holds the writer lock.
func (w *writer) reserve(t *tree, n uint64) (run, error) {
	// Keep the reservations in one space.
	sp := w.sp
	if t.sp == nil {
		t.sp = sp
	}
	if t.sp != sp || w.closing.Load() {
		return run{}, errDropped
	}

	// Take n pages from the lowest free run holding them, else the largest
	// free run, else from the file end. A checkpoint rewrites much of the
	// index, so asking for a write batch would put each one at the file end
	// once values have cut the free space into shorter runs.
	r, ok := sp.fit(1, n)
	if !ok {
		r = sp.alloc(n)
	}
	sp.reserved = append(sp.reserved, r)
	return r, nil
}

// carryOver collects the overlay entries that follow t's state, so
// saveTree folds in under the writer lock only the commits published
// during the fold. Records another process appended reload the space,
// which drops the tree, so this handle's records are the only ones to
// follow.
func (w *writer) carryOver(ctx context.Context, t *tree) error {
	// Scan the published state without the writer lock: published states
	// do not change. The filter is sized for the current tree, which the
	// new one replaces at about the same size.
	pre := w.db.cur.Load()
	t.carried, t.carriedSeq = newCarried(t.st.seq, w.db.opts.checkpointLimit(pre.sb)), pre.seq
	pre.overlay.Scan(t.carried.add)

	// Fold in the records published meanwhile, taking the writer lock only
	// to copy their list. A record is never changed once listed. Folding
	// again would chase a writer that commits as fast as the fold runs.
	if err := w.lock(ctx); err != nil {
		return err
	}
	changed := w.changed
	w.unlock(nil)
	t.fold(changed)
	return nil
}

// fold adds the changes of the records in changed after the carried ones
// to t.carried.
func (t *tree) fold(changed []*record) {
	for _, r := range changed {
		if r.seq > t.carriedSeq {
			t.carried.addRecord(r)
			t.carriedSeq = r.seq
		}
	}
}

// saveTree saves t and a superblock naming it, and publishes the state
// moved onto it. Commits after t's state stay in the overlay, and the space
// is saved as of the last of them. A tree built on a replaced checkpoint is
// dropped. The caller holds the writer lock.
func (w *writer) saveTree(ctx context.Context, t *tree) error {
	// End the reservation. Drop a tree whose space was reloaded or whose
	// checkpoint another replaced, freeing its pages at once and evicting
	// them from the cache: no reader has seen them.
	db, sp, b, st := w.db, w.sp, t.b, t.st
	cur := db.cur.Load()
	if t.sp != nil && t.sp != sp {
		t.evict()
		return nil
	}
	out := b.out
	sp.reserved = slices.DeleteFunc(sp.reserved, func(r run) bool { return slices.Contains(out.runs, r) })
	if cur.gen != st.gen {
		t.evict()
		for _, r := range out.runs {
			sp.addFree(r)
		}
		return nil
	}
	sp.addFree(out.unused())

	// The root is the new top page, or an existing page when nothing above
	// it changed.
	var root uint64
	if t.top != nil {
		root = t.top.page
	}
	fresh := out.count

	// Free the replaced pages, the old space record, and the log chunks
	// before the one holding the tree's next record once no snapshot reads
	// the previous tree.
	sp.freePages(b.freed, st.seq)
	old := cur.sb.space
	sp.freePages(pageRange(old.off/pageSize, pagesFor(int(old.n))), st.seq)
	keep := slices.IndexFunc(sp.chunks, func(c run) bool { return c.end()*pageSize == st.end })
	if keep < 0 {
		return errors.Wrap(ErrCorrupt, "log chunk not in the space")
	}
	for _, c := range sp.chunks[:keep] {
		sp.freePages(pageRange(c.start, c.n), st.seq)
	}
	sp.chunks = sp.chunks[keep:]

	// Close the open value extent and save the space.
	sp.closeOpen(cur.seq)
	ref, err := w.writeSpace(sp)
	if err != nil {
		return err
	}

	// Order the pages and space before the superblock that names them. A
	// durable barrier also makes the records and the previous superblock
	// durable.
	prev := db.flush.newest()
	durable, err := db.s.flushBarrier()
	if err != nil {
		return err
	}
	if durable {
		db.flush.setDurable(cur, prev)
	}

	// Switch superblocks. The new one becomes durable with a later flush,
	// which releases the pages it replaced.
	sb := superblock{
		gen:       cur.gen + 1,
		seq:       st.seq,
		root:      root,
		count:     uint64(st.count),                                        // #nosec G115 -- key counts are not negative.
		treePages: cur.sb.treePages + uint64(fresh) - uint64(len(b.freed)), // #nosec G115 -- counts are not negative.
		space:     ref,
		spaceSeq:  cur.seq,
		logPos:    st.pos,
		logEnd:    st.end,
		logCrc:    st.crc,
	}
	db.ownGen.Store(sb.gen)
	if _, err := db.s.WriteAt(sb.encode(), sb.page()*pageSize); err != nil {
		return err
	}

	// Publish the tree under the commits after it: those carried over
	// before the writer lock, then those published since.
	if t.carried == nil {
		t.carried, t.carriedSeq = newCarried(st.seq, db.opts.checkpointLimit(sb)), cur.seq
		cur.overlay.Scan(t.carried.add)
	}
	t.fold(w.changed)
	w.changed = nil
	if err := db.publish(ctx, cur.onto(sb, t.carried, cur.logBytes-st.logBytes)); err != nil {
		return err
	}
	w.spGen = sb.gen
	return w.punch()
}

// pageRange lists n pages from start.
func pageRange(start, n uint64) []uint64 {
	pages := make([]uint64, n)
	for i := range pages {
		pages[i] = start + uint64(i)
	}
	return pages
}

// writeSpace saves sp in newly allocated pages.
func (w *writer) writeSpace(sp *space) (extentRef, error) {
	r, b := sp.allocRecord()
	if _, err := w.db.s.WriteAt(b, pageOff(r.start)); err != nil {
		return extentRef{}, err
	}
	return extentRef{off: r.start * pageSize, n: uint32(len(b)), crc: checksum(b)}, nil // #nosec G115 -- the space record is far below 4 GiB.
}

// relocate moves up to budget value bytes out of sparse pages, and out of
// the file's last quarter when free space below would take them, reading at
// most visits entries from where the last pass stopped. The caller holds
// the writer lock.
func (w *writer) relocate(ctx context.Context, budget int64, visits int) error {
	// Start the walk at the saved key.
	db, sp := w.db, w.sp
	st := db.cur.Load()
	it := newIterator(db.p, st, nil, nil, false)
	if w.relocKey != nil {
		_ = it.Seek(w.relocKey)
	} else {
		it.Next()
	}

	// Collect values in sparse pages or in the last quarter of the file.
	tail := sp.end * 3 / 4
	var moves []tentry
	var moved int64
	for ; it.Valid() && moved < budget && visits > 0; it.Next() {
		visits--
		v := it.cur.val
		if !v.isRef {
			continue
		}
		p := v.ref.off / pageSize
		sparse := sp.live[p] < pageSize/2
		if !sparse && (p < tail || !sp.lower(int(v.ref.n), p)) {
			continue
		}
		data, err := db.p.readValue(v)
		if err != nil {
			return err
		}
		moves = append(moves, tentry{key: bytes.Clone(it.Key()), data: data})
		moved += int64(len(data))
	}
	if err := it.Err(); err != nil {
		return err
	}

	// Save where the next walk resumes.
	w.relocKey = nil
	if it.Valid() {
		w.relocKey = bytes.Clone(it.Key())
	}
	if len(moves) == 0 {
		return nil
	}

	// Rewrite the values: placement prefers the lowest free pages.
	sp.closeOpen(st.seq)
	return w.commit(ctx, st, moves, true)
}

// afterCommit runs checkpoint work. Values leave sparse pages in the
// background compactor, outside writers' commits. The caller holds the
// writer lock.
func (w *writer) afterCommit(ctx context.Context) error {
	// Report a failed background checkpoint, and start one when the
	// overlay or the log after the last one has grown enough. Rewrites of
	// the same keys grow the log but not the overlay.
	db := w.db
	if err := w.ckptErr; err != nil {
		w.ckptErr = nil
		return err
	}
	st := db.cur.Load()
	if limit := db.opts.checkpointLimit(st.sb); st.overlayBytes > limit || st.logBytes > 2*limit {
		w.startCheckpoint()
	}
	return nil
}

// compact moves every value out of sparse pages and the file's tail,
// checkpoints, and releases freed space. The caller holds the writer lock.
func (w *writer) compact(ctx context.Context) error {
	// Each round moves values in one full pass of bounded batches,
	// checkpoints to free the replaced pages, and moves the log into the
	// space they leave. The second round fills the space the first
	// released and frees the old log chunk.
	for range 2 {
		w.relocKey = nil
		for first := true; first || w.relocKey != nil; first = false {
			if err := w.relocate(ctx, compactBatch, math.MaxInt); err != nil {
				return err
			}
		}
		if err := w.checkpointDurable(ctx); err != nil {
			return err
		}
		if err := w.releaseNow(ctx); err != nil {
			return err
		}
		if err := w.moveLog(ctx); err != nil {
			return err
		}
	}
	return nil
}

// moveLog continues the log in the lowest free run below its current chunk,
// so the next checkpoint frees the chunk and the file can shrink. The caller
// holds the writer lock.
func (w *writer) moveLog(ctx context.Context) error {
	// Take the lowest free chunk, keeping the log where it is when the chunk
	// lies above it.
	db, sp := w.db, w.sp
	st := db.cur.Load()
	chunk, ok := sp.fit(extentPages, extentPages)
	if !ok {
		return nil
	}
	if chunk.start*pageSize > st.pos {
		sp.addFree(chunk)
		return nil
	}

	// Link the log to the chunk.
	sp.chunks = append(sp.chunks, chunk)
	link := newLink(st.seq+1, chunk)
	if _, err := db.s.WriteAt(link.encode(st.crc), fileOff(st.pos)); err != nil {
		return err
	}
	if err := db.s.flushOrdered(); err != nil {
		return err
	}

	// Publish the state that continues in the chunk.
	next := *st
	next.seq, next.crc = next.seq+1, link.crc
	next.pos, next.end = chunk.start*pageSize, chunk.end()*pageSize
	next.refs = nil
	if err := db.publish(ctx, &next); err != nil {
		return err
	}
	w.spSeq = next.seq
	return nil
}

// checkpointDurable checkpoints and flushes, so releaseNow can return the
// pages the checkpoint replaced. The caller holds the writer lock.
func (w *writer) checkpointDurable(ctx context.Context) error {
	if err := w.checkpoint(ctx); err != nil {
		return err
	}
	st := w.db.cur.Load()
	return w.db.flush.syncThrough(ctx, st.seq, st.ckpt)
}

// releaseNow returns freed space no snapshot reads to the file system
// without writing a record. The next commit record carries release marks at
// least as high, so recovery repeats the release before reusing the space.
// In a shared file, another process's background checkpoint may have
// written pages it reserved from the released runs or past the space end,
// and would not notice a punch that leaves no record, so the punch needs
// the build lock: held by this handle's build, or taken here when free. The
// caller holds the writer lock.
func (w *writer) releaseNow(ctx context.Context) error {
	// Release what no snapshot reads.
	if err := w.releaseSpace(ctx, w.db.cur.Load(), &record{}); err != nil {
		return err
	}

	// Punch when no other process's checkpoint can hold the pages.
	if w.db.opts.SingleProcess || w.buildLocked.Load() {
		return w.punch()
	}

	// Else take the build lock first, leaving the released runs queued
	// for the next punch while another process builds.
	ok, err := w.db.s.lock(lockBuild, false)
	if err != nil || !ok {
		return err
	}
	defer func() { _ = w.db.s.unlock(lockBuild) }()
	return w.punch()
}
