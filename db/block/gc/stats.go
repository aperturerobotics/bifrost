package block_gc

import "time"

// Stats holds GC cycle statistics.
type Stats struct {
	// AtomicSweepCount is the number of candidates atomic sweeps rechecked.
	AtomicSweepCount int
	// AtomicSweepDuration is time spent in atomic sweeps, including rechecks.
	AtomicSweepDuration time.Duration
	// NodesSwept is the number of nodes swept.
	NodesSwept int
	// UnreferencedNodeCount is the number of unreferenced node entries read.
	UnreferencedNodeCount int
	// Candidates is the number of unreferenced nodes the last pass listed.
	Candidates int
	// More reports that a budgeted collection stopped with candidates left.
	More bool
	// RemoveNodeRefsCount is the number of nodes whose outgoing refs were removed.
	RemoveNodeRefsCount int
	// RemoveUnreferencedEdgeCount is the number of unreferenced marker edges removed.
	RemoveUnreferencedEdgeCount int
	// OnSweptCount is the number of onSwept callbacks run.
	OnSweptCount int
	// RemoveBlockCount is the number of physical block deletes attempted.
	RemoveBlockCount int
	// Duration is how long the GC cycle took.
	Duration time.Duration
	// UnreferencedScanDuration is time spent listing unreferenced nodes.
	UnreferencedScanDuration time.Duration
	// RemoveNodeRefsDuration is time spent removing outgoing node refs.
	RemoveNodeRefsDuration time.Duration
	// RemoveUnreferencedEdgeDuration is time spent removing unreferenced markers.
	RemoveUnreferencedEdgeDuration time.Duration
	// OnSweptDuration is time spent in onSwept callbacks.
	OnSweptDuration time.Duration
	// RemoveBlockDuration is time spent physically deleting block-backed nodes.
	RemoveBlockDuration time.Duration
	// BlockPasses is the number of block passes that marked the graph.
	BlockPasses int
	// MarkDuration is time spent marking blocks.
	MarkDuration time.Duration
	// MarkScans is the number of child list scans the last mark made, zero
	// for a walk.
	MarkScans int
	// MarkLongestRead is the last mark's longest read transaction, zero for
	// a walk.
	MarkLongestRead time.Duration
	// LiveNodes is the number of nodes the last mark reached.
	LiveNodes int
	// UnreachedBlocks is the number of tracked blocks the last mark did not
	// reach.
	UnreachedBlocks int
	// UnknownBlocks is the number of reached blocks of unknown references the
	// last mark found. A pass that finds any sweeps no block.
	UnknownBlocks int
	// UnknownSample lists some of those blocks.
	UnknownSample []string
	// RescuedNodes is the number of nodes sweeps rescued from the adopt log.
	RescuedNodes int
	// BlocksSwept is the number of blocks block passes removed.
	BlocksSwept int
}

// RecordMark records a block pass's mark.
func (s *Stats) RecordMark(mark *BlockMark) {
	// Count the pass and copy the mark's totals.
	s.BlockPasses++
	s.LiveNodes = mark.Live.Len()
	s.UnreachedBlocks = len(mark.Unreached)
	s.UnknownBlocks = mark.Unknown
	s.UnknownSample = mark.UnknownSample
	s.MarkScans = mark.Scans
	s.MarkLongestRead = mark.LongestRead
}
