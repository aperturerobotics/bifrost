package kvtx

import "context"

// PrefetchTxOps loads the index of keys ahead of reads in a transaction.
type PrefetchTxOps interface {
	// Prefetch loads the index records of keys, reading no values, so later
	// reads of the keys in this transaction need no index read of their own.
	// A store implements it when it can read many keys' index together
	// faster than one at a time.
	Prefetch(ctx context.Context, keys [][]byte) error
}

// Prefetch calls PrefetchTxOps.Prefetch when ops implements it and otherwise
// does nothing, since reading the keys ahead would only repeat each read.
func Prefetch(ctx context.Context, ops TxOps, keys [][]byte) error {
	if p, ok := ops.(PrefetchTxOps); ok && len(keys) != 0 {
		return p.Prefetch(ctx, keys)
	}
	return nil
}
