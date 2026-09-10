package provider_spacewave

import (
	"context"

	"github.com/aperturerobotics/util/csync"
	"github.com/s4wave/spacewave/db/kvtx"
)

// blockPublicationRetention belongs to the local block cache. Removing cached
// bytes invalidates its graph proofs before deletion can become durable.
type blockPublicationRetention struct {
	mu    csync.Mutex
	store kvtx.Store
}

// invalidate excludes graph retention until the caller finishes deleting bytes.
// A failure leaves the bytes intact. A crash after invalidation requires a fresh
// graph walk, even when the physical deletion never completed.
func (r *blockPublicationRetention) invalidate(ctx context.Context) (func(), error) {
	if r == nil {
		return func() {}, nil
	}
	release, err := r.mu.Lock(ctx)
	if err != nil {
		return nil, err
	}
	err = kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) {
		return r.store.NewTransaction(ctx, true)
	}, func(ctx context.Context, tx kvtx.Tx) error {
		return tx.ScanPrefixKeys(ctx, nil, func(key []byte) error { return tx.Delete(ctx, key) })
	})
	if err != nil {
		release()
		return nil, err
	}
	return release, nil
}
