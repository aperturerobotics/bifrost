package provider_spacewave

import (
	"bytes"
	"context"
	"encoding/hex"

	"github.com/pkg/errors"
	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/kvtx"
)

// verifiedSOStateCacheKeyPrefix is the ObjectStore key prefix for verified SO cache state.
const verifiedSOStateCacheKeyPrefix = "verified-so-state/"

// verifiedSOHistoryKeyPrefix is the ObjectStore key prefix for retained
// verified config changes. Each change is stored once under its hash.
const verifiedSOHistoryKeyPrefix = "verified-so-history/"

// verifiedSOStateCacheKey returns the ObjectStore key for a verified SO cache entry.
func verifiedSOStateCacheKey(soID string) []byte {
	return []byte(verifiedSOStateCacheKeyPrefix + soID)
}

// verifiedSOHistoryPrefix returns the ObjectStore key prefix of soID's retained history.
func verifiedSOHistoryPrefix(soID string) []byte {
	return []byte(verifiedSOHistoryKeyPrefix + soID + "/")
}

// verifiedSOHistoryKey returns the ObjectStore key of one retained config change.
func verifiedSOHistoryKey(soID string, hash []byte) []byte {
	return append(verifiedSOHistoryPrefix(soID), hex.EncodeToString(hash)...)
}

// writeVerifiedSOStateCache serializes verified SO cache state to the account
// ObjectStore and retains changes in the same transaction. Changes already
// retained are not written again.
func (a *ProviderAccount) writeVerifiedSOStateCache(
	ctx context.Context,
	soID string,
	cache *api.VerifiedSOStateCache,
	changes []*sobject.SOConfigChange,
) error {
	// Validate the store and arguments.
	if a.objStore == nil {
		return errors.New("account object store not ready")
	}
	if soID == "" {
		return errors.New("shared object id is required")
	}
	if cache == nil {
		return errors.New("verified SO cache is required")
	}

	// Write the state record and its new changes in one transaction.
	data, err := cache.MarshalVT()
	if err != nil {
		return errors.Wrap(err, "marshal verified SO cache")
	}
	err = kvtx.RunTransaction(ctx, true,
		func(ctx context.Context) (kvtx.Tx, error) {
			return a.objStore.NewTransaction(ctx, true)
		},
		func(ctx context.Context, tx kvtx.Tx) error {
			// Write the verified state record.
			if err := tx.Set(ctx, verifiedSOStateCacheKey(soID), data); err != nil {
				return errors.Wrap(err, "set verified SO cache")
			}

			// Append each change not yet retained under its hash.
			for _, change := range changes {
				if err := writeVerifiedSOHistoryEntry(ctx, tx, soID, change); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return errors.Wrap(err, "write verified SO cache")
	}
	a.invalidateSelfEnrollmentEntry(soID)
	return nil
}

// writeVerifiedSOHistoryEntry stores change under its hash unless it is already retained.
func writeVerifiedSOHistoryEntry(ctx context.Context, tx kvtx.Tx, soID string, change *sobject.SOConfigChange) error {
	// Address the change by its hash and skip it when already retained.
	hash, err := sobject.HashSOConfigChange(change)
	if err != nil {
		return errors.Wrap(err, "hash verified SO history entry")
	}
	key := verifiedSOHistoryKey(soID, hash)
	found, err := tx.Exists(ctx, key)
	if err != nil || found {
		return err
	}

	// Store the serialized change.
	data, err := change.MarshalVT()
	if err != nil {
		return errors.Wrap(err, "marshal verified SO history entry")
	}
	if err := tx.Set(ctx, key, data); err != nil {
		return errors.Wrap(err, "set verified SO history entry")
	}
	return nil
}

// readVerifiedSOHistoryEntry returns the retained config change of soID with
// hash, or nil when it is not retained.
func (a *ProviderAccount) readVerifiedSOHistoryEntry(
	ctx context.Context,
	soID string,
	hash []byte,
) (*sobject.SOConfigChange, error) {
	// Validate the store.
	if a.objStore == nil {
		return nil, errors.New("account object store not ready")
	}

	// Read and decode the entry, leaving it nil when absent.
	var entry *sobject.SOConfigChange
	err := kvtx.RunTransaction(ctx, false,
		func(ctx context.Context) (kvtx.Tx, error) {
			return a.objStore.NewTransaction(ctx, false)
		},
		func(ctx context.Context, tx kvtx.Tx) error {
			// Reset the result so a retried attempt starts clean.
			entry = nil
			data, found, err := tx.Get(ctx, verifiedSOHistoryKey(soID, hash))
			if err != nil || !found {
				return err
			}
			entry = &sobject.SOConfigChange{}
			return entry.UnmarshalVT(data)
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "read verified SO history entry")
	}
	return entry, nil
}

// deleteVerifiedSOStateCache removes the persisted verified SO cache entry
// and the retained history for soID, if any. Missing entries are not an
// error: the caller is asking for a cold next mount, which a missing entry
// already produces.
func (a *ProviderAccount) deleteVerifiedSOStateCache(
	ctx context.Context,
	soID string,
) error {
	// Validate the store and arguments.
	if a.objStore == nil {
		return errors.New("account object store not ready")
	}
	if soID == "" {
		return errors.New("shared object id is required")
	}

	// Delete the state record and its history in one transaction.
	err := kvtx.RunTransaction(ctx, true,
		func(ctx context.Context) (kvtx.Tx, error) {
			return a.objStore.NewTransaction(ctx, true)
		},
		func(ctx context.Context, tx kvtx.Tx) error {
			// Delete the verified state record.
			if err := tx.Delete(ctx, verifiedSOStateCacheKey(soID)); err != nil {
				return errors.Wrap(err, "delete verified SO cache")
			}

			// Collect the retained history keys, then delete them.
			var keys [][]byte
			err := tx.ScanPrefixKeys(ctx, verifiedSOHistoryPrefix(soID), func(key []byte) error {
				keys = append(keys, bytes.Clone(key))
				return nil
			})
			if err != nil {
				return errors.Wrap(err, "list verified SO history")
			}
			for _, key := range keys {
				if err := tx.Delete(ctx, key); err != nil {
					return errors.Wrap(err, "delete verified SO history")
				}
			}
			return nil
		},
	)
	if err != nil {
		return errors.Wrap(err, "delete verified SO cache")
	}
	a.invalidateSelfEnrollmentEntry(soID)
	return nil
}

// InvalidateVerifiedChain clears the persisted verified config-chain record
// for soID so the next mount re-verifies from scratch via /config-chain.
// Used by rejoin, recovery, and forced re-verification flows where the
// previously trusted chain head is no longer authoritative. The live
// cloudSOHost (if any) keeps its in-memory verified state for the
// remainder of its session; only future mounts observe the cold cache.
func (a *ProviderAccount) InvalidateVerifiedChain(ctx context.Context, soID string) error {
	return a.deleteVerifiedSOStateCache(ctx, soID)
}

// loadVerifiedSOStateCache reads verified SO cache state from the account ObjectStore.
func (a *ProviderAccount) loadVerifiedSOStateCache(
	ctx context.Context,
	soID string,
) (*api.VerifiedSOStateCache, error) {
	// Validate the store and arguments.
	if a.objStore == nil {
		return nil, errors.New("account object store not ready")
	}
	if soID == "" {
		return nil, errors.New("shared object id is required")
	}

	// Read and decode the record, leaving it nil when absent.
	var cache *api.VerifiedSOStateCache
	err := kvtx.RunTransaction(ctx, false,
		func(ctx context.Context) (kvtx.Tx, error) {
			return a.objStore.NewTransaction(ctx, false)
		},
		func(ctx context.Context, tx kvtx.Tx) error {
			// Reset the result so a retried attempt starts clean.
			cache = nil
			data, found, err := tx.Get(ctx, verifiedSOStateCacheKey(soID))
			if err != nil {
				return errors.Wrap(err, "get verified SO cache")
			}
			if !found {
				return nil
			}

			// Decode the record.
			next := &api.VerifiedSOStateCache{}
			if err := next.UnmarshalVT(data); err != nil {
				return errors.Wrap(err, "unmarshal verified SO cache")
			}
			cache = next
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "open read transaction")
	}
	return cache, nil
}
