package provider_local

import (
	"context"
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/kvtx"
)

// SOConfigHistoryCheckpointKey addresses the locally held configuration from
// which retained history begins. Its value is a serialized SharedObjectConfig.
func SOConfigHistoryCheckpointKey(sharedObjectID string) []byte {
	return []byte("so/" + sharedObjectID + "/config-history/checkpoint")
}

// SOConfigHistoryEntryKey addresses a serialized SOConfigChange by its hash.
func SOConfigHistoryEntryKey(sharedObjectID string, hash []byte) []byte {
	return []byte("so/" + sharedObjectID + "/config-history/entry/" + hex.EncodeToString(hash))
}

// WriteSOConfigHistory retains verified transitions in the transaction that
// writes their resulting SOState. The caller commits both or discards both.
// Ordinary state writes do not establish or replace a history checkpoint.
// Retention records held trust; it does not authenticate legacy state or admit peers.
func WriteSOConfigHistory(
	ctx context.Context,
	tx kvtx.Tx,
	sharedObjectID string,
	current, next *sobject.SharedObjectConfig,
	changes []*sobject.SOConfigChange,
) error {
	// Ordinary writes leave retained history intact, including legacy history gaps.
	if len(changes) == 0 {
		return nil
	}
	if err := current.Validate(); err != nil {
		return errors.Wrap(err, "config history checkpoint")
	}

	// Verify every transition against the state held under the provider lock.
	checkpoint := current.CloneVT()
	for _, change := range changes {
		accepted, err := sobject.VerifyConfigChange(current, change)
		if err != nil {
			return err
		}
		if err := accepted.Validate(); err != nil {
			return errors.Wrap(err, "config history transition")
		}
		if len(checkpoint.GetConfigChainHash()) == 0 {
			checkpoint = accepted.CloneVT()
		}
		current = accepted
	}
	if !current.EqualVT(next) {
		return errors.New("config history does not match written state")
	}

	// Preserve the first locally held checkpoint across later writes and gaps.
	checkpointKey := SOConfigHistoryCheckpointKey(sharedObjectID)
	_, found, err := tx.Get(ctx, checkpointKey)
	if err != nil {
		return err
	}
	if !found {
		data, err := checkpoint.MarshalVT()
		if err != nil {
			return err
		}
		if err := tx.Set(ctx, checkpointKey, data); err != nil {
			return err
		}
	}

	// Store entries separately so opening state does not load the full history.
	for _, change := range changes {
		hash, err := sobject.HashSOConfigChange(change)
		if err != nil {
			return err
		}
		data, err := change.MarshalVT()
		if err != nil {
			return err
		}
		if err := tx.Set(ctx, SOConfigHistoryEntryKey(sharedObjectID, hash), data); err != nil {
			return err
		}
	}
	return nil
}
