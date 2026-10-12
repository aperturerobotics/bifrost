package s4db

import (
	"bytes"
)

// placeBatch is the number of adjacent pages written at once.
const placeBatch = 256

// reserveBatch is the number of pages a checkpoint reserves at once. Each
// reservation takes the writer lock, and the unused tail of the last one
// returns to the free runs when the tree is saved or dropped.
const reserveBatch = 4096

// pageWriter numbers a checkpoint's new pages in the order it receives them
// and writes them in batches of adjacent pages, so a checkpoint holds the
// encoding of at most one batch. It caches each page decoded from its own
// copy, which holds no buffer of the pages it replaced.
type pageWriter struct {
	// reserve returns a run of at least one and at most n unused pages.
	reserve func(n uint64) (run, error)
	// write writes encoded pages from page.
	write func(buf []byte, page uint64) error
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

// add writes d to the next page and returns the page and its decoded copy.
func (w *pageWriter) add(d *draft) (uint64, *node, error) {
	// Reserve more pages when the last run is used up.
	if l := len(w.runs); l == 0 || w.next == w.runs[l-1].end() {
		if err := w.flush(); err != nil {
			return 0, nil, err
		}
		r, err := w.reserve(reserveBatch)
		if err != nil {
			return 0, nil, err
		}
		w.runs = append(w.runs, r)
		w.next, w.at = r.start, r.start
	}
	page := w.next
	w.next++
	w.count++

	// Encode the page and cache its decoded copy.
	if w.buf == nil {
		w.buf = make([]byte, 0, placeBatch*pageSize)
	}
	w.buf = d.encode(w.buf)
	n, err := decodeNode(bytes.Clone(w.buf[len(w.buf)-pageSize:]))
	if err != nil {
		return 0, nil, err
	}
	w.cache.put(page, n)

	// Write a full batch.
	if len(w.buf) == cap(w.buf) {
		err = w.flush()
	}
	return page, n, err
}

// flush writes the pages in buf.
func (w *pageWriter) flush() error {
	// Write the buffered pages, if any, and advance past them.
	if len(w.buf) == 0 {
		return nil
	}
	err := w.write(w.buf, w.at)
	w.at += uint64(len(w.buf) / pageSize)
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
