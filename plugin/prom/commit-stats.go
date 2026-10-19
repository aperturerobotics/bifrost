package prom_plugin

// CommitStats reports what one commit wrote.
type CommitStats struct {
	// Segments is the number of segments written, one per store with samples.
	Segments int
	// Samples is the number of samples in the written segments.
	Samples int
	// Blocks is the number of encoded blocks the transaction submitted for
	// storage, including key-value tree nodes and the World's own blocks.
	Blocks uint64
	// Bytes is the number of encoded bytes of those blocks.
	Bytes uint64
}
