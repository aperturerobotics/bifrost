package common

// Pragmas configures tunable SQLite pragmas applied during Open.
// A zero value leaves SQLite's compiled default in place.
type Pragmas struct {
	// FullSync selects synchronous=FULL; the existing default is NORMAL.
	// Connectors that open several physical connections must configure each one.
	FullSync bool
	// CacheSize sets the SQLite cache_size pragma. Positive = pages,
	// negative = KiB. 0 leaves the SQLite default.
	CacheSize int32
	// MmapSize sets the SQLite mmap_size pragma in bytes. 0 leaves the
	// SQLite default (mmap disabled).
	MmapSize int64
	// TempStore sets the SQLite temp_store pragma. Valid values are 0
	// (DEFAULT), 1 (FILE), 2 (MEMORY). 0 leaves the SQLite default.
	TempStore int32
	// PageSize sets the SQLite page_size pragma in bytes. Must be a power
	// of two between 512 and 65536. 0 leaves the SQLite default. Only
	// effective on a fresh database.
	PageSize int32
}
