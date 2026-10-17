package s4db

import (
	"bytes"
)

// placeBatch is the shortest free run a checkpoint reserves from before it
// takes pages from the file end.
const placeBatch = 256

// reserveBatch is the number of pages a checkpoint reserves at once. The
// unused tail of the last run returns to the free runs when the tree is
// saved or dropped.
const reserveBatch = 4096

// pageWriter numbers a checkpoint's new pages in the order it receives them
// and writes each reserved run once it is full, so a checkpoint holds the
// encoding of at most one run. It caches each page decoded from its own
// copy, which holds no buffer of the pages it replaced.
type pageWriter struct {
	// turn writes the encoded pages in buf from page at, then reserves a run
	// of at least one and at most n unused pages when n is not zero. A
	// background checkpoint takes the writer lock once per turn, so filling
	// whole runs keeps its turns few while writers queue for the lock.
	turn func(buf []byte, at, n uint64) (run, error)
	// cache receives the decoded pages.
	cache *cache
	// runs lists the reserved runs in order.
	runs []run
	// next is the next unused page of the last run.
	next uint64
	// buf holds the encoded pages not yet written, from page at.
	buf []byte
	// at is the page of the first page in buf.
	at uint64
	// count is the number of pages added.
	count int
}

// add writes d to the next page and returns the page.
func (w *pageWriter) add(d *draft) (uint64, error) {
	// Write the full run and reserve the next one in one turn.
	if l := len(w.runs); l == 0 || w.next == w.runs[l-1].end() {
		r, err := w.turn(w.buf, w.at, reserveBatch)
		w.buf = w.buf[:0]
		if err != nil {
			return 0, err
		}
		w.runs = append(w.runs, r)
		w.next, w.at = r.start, r.start
		if size := int(r.n) * pageSize; cap(w.buf) < size { // #nosec G115 -- a run holds at most reserveBatch pages.
			w.buf = make([]byte, 0, size)
		}
	}
	page := w.next
	w.next++
	w.count++

	// Encode the page and cache its decoded copy.
	w.buf = d.encode(w.buf)
	n, err := decodeNode(bytes.Clone(w.buf[len(w.buf)-pageSize:]))
	if err != nil {
		return 0, err
	}
	w.cache.put(page, n)
	return page, nil
}

// flush writes the pages of the last run.
func (w *pageWriter) flush() error {
	// Write the buffered pages, if any, without reserving more.
	if len(w.buf) == 0 {
		return nil
	}
	_, err := w.turn(w.buf, w.at, 0)
	w.buf = w.buf[:0]
	return err
}

// unused returns the reserved pages no page was added to.
func (w *pageWriter) unused() run {
	l := len(w.runs)
	if l == 0 {
		return run{}
	}
	return run{start: w.next, n: w.runs[l-1].end() - w.next}
}
