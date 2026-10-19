package block_rpc_server

// ReadPolicy is the policy a BlockStore service applies to every read it
// serves, on top of the class each request carries.
type ReadPolicy struct {
	// ReadAheadBytes is the range-fetch minimum, or zero for none.
	ReadAheadBytes int
	// Background marks every read as background work, for a service that
	// exists only to serve a background copy.
	Background bool
}
