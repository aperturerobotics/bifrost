// Package prom_plugin stores Prometheus samples in a World and answers PromQL
// over them.
//
// Each scrape target is one World object. Its root is a block-backed key-value
// tree holding the target's labels, its series table, and one immutable
// segment per commit window. A Store appends samples to an in-memory window
// and commits the window as one segment, and a Queryable serves the fork's
// storage.Queryable interface from the committed segments of several targets.
package prom_plugin

// ChunkSamples is the number of samples after which a window starts a new
// chunk for a series, the same cut Prometheus uses for its head chunks.
const ChunkSamples = 120
