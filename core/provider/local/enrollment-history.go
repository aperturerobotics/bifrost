package provider_local

import (
	"bytes"
	"context"

	"github.com/s4wave/spacewave/core/pairing"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/kvtx"
)

// retainEnrollmentHistory keeps the authenticated checkpoint's earlier lineage
// so this replica can prove later account changes to an offline participant.
func (a *ProviderAccount) retainEnrollmentHistory(ctx context.Context, checkpoint *pairing.SharedObject) error {
	base := checkpoint.GetHistoryBase()
	if base == nil {
		return nil
	}
	next := checkpoint.GetState().GetConfig()
	if err := sobject.VerifyConfigChainSuffix(base, next, checkpoint.GetHistory()); err != nil {
		return err
	}
	object, release, err := a.MountSharedObject(ctx, checkpoint.GetEntry().GetRef(), nil)
	if err != nil {
		return err
	}
	defer release()
	local := object.(*SharedObject)
	return kvtx.RunTransaction(ctx, true, func(ctx context.Context) (kvtx.Tx, error) {
		return local.objStore.NewTransaction(ctx, true)
	}, func(ctx context.Context, tx kvtx.Tx) error {
		if genesis := checkpoint.GetGenesis(); genesis != nil {
			if err := sobject.VerifyConfigChain([]*sobject.SOConfigChange{genesis}); err != nil {
				return err
			}
			hash, err := sobject.HashSOConfigChange(genesis)
			if err != nil {
				return err
			}
			if !bytes.Equal(hash, base.GetConfigChainHash()) {
				return sobject.ErrConfigHistoryUnavailable
			}
			data, err := genesis.MarshalVT()
			if err != nil {
				return err
			}
			if err := tx.Set(ctx, SOConfigHistoryEntryKey(local.GetSharedObjectID(), hash), data); err != nil {
				return err
			}
		}
		if err := WriteSOConfigHistory(ctx, tx, local.GetSharedObjectID(), base, next, checkpoint.GetHistory()); err != nil {
			return err
		}
		return WriteSOConfigCheckpoint(ctx, tx, local.GetSharedObjectID(), base, true)
	})
}
