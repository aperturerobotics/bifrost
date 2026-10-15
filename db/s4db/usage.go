package s4db

import "context"

// Usage describes how the file's pages are used.
type Usage struct {
	// FileBytes is the file length.
	FileBytes int64
	// FreeBytes is the length of the free runs. They are punched in batches,
	// so they take little disk.
	FreeBytes int64
	// ValueBytes is the total length of the live values stored outside the
	// index.
	ValueBytes int64
	// ValuePageBytes is the length of the pages holding live values. The
	// difference from ValueBytes is the dead space inside them.
	ValuePageBytes int64
	// IndexKeys is the number of keys in the checkpoint's index.
	IndexKeys int64
	// IndexBytes is the length of the checkpoint's index pages.
	IndexBytes int64
	// LogBytes is the length of the log chunks.
	LogBytes int64
	// PendingBytes is the length of the freed runs a snapshot may still
	// read.
	PendingBytes int64
}

// Usage reports how the file's pages are used as of the latest commit. It
// takes the writer lock.
func (db *DB) Usage(ctx context.Context) (Usage, error) {
	// Hold the writer lock so the space is current and stable.
	if err := db.w.lock(ctx); err != nil {
		return Usage{}, err
	}

	// Read the space, then the checkpoint's index.
	u := db.w.sp.usage()
	sb := db.cur.Load().sb
	u.IndexKeys = int64(sb.count) // #nosec G115 -- key counts fit in int64.
	u.IndexBytes = pageOff(sb.treePages)
	db.w.unlock(nil)
	return u, nil
}
