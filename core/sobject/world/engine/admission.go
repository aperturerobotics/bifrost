package sobject_world_engine

import (
	"context"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
)

// writeAdmission is a local write admitted to a publication whose receipt the
// engine has not settled. It keeps what was durable before the publication,
// which a failed receipt restores.
type writeAdmission struct {
	// receipt resolves with the publication's durable result. The write owns
	// it; the engine only reads it.
	receipt *block.PublicationReceipt
	// localID is the local id of the admitted operation, or empty when the
	// SharedObject did not report it.
	localID string
	// base is the replay's checkpoint World when the write was admitted. The
	// saved span holds positions above it.
	base *InnerState
	// span and spanLen are the saved replay span before the publication.
	span    *block.BlockRef
	spanLen int
	// retained is the head held under the accepted World root before the
	// publication.
	retained *bucket.ObjectRef
}

// admit records a publication admitted with receipt, then records its replay
// save and its World as held, so the writes after it fork from that World and
// chain their publications after it. The caller holds the writer lock and
// passes the World the publication holds under the accepted World root.
func (e *soEngine) admit(receipt *block.PublicationReceipt, held *bucket.ObjectRef) {
	e.admissions = append(e.admissions, &writeAdmission{
		receipt:  receipt,
		base:     e.replay.base,
		span:     e.replay.span,
		spanLen:  e.replay.spanLen,
		retained: e.retained,
	})
	e.retained = held
	e.replay.saved()
}

// lastReceipt returns the receipt of the last admitted publication, which the
// next publication follows, or nil when every admission settled. A failed
// receipt is returned too, so a write forked from its World fails with it. The
// caller holds the writer lock.
func (e *soEngine) lastReceipt() *block.PublicationReceipt {
	if len(e.admissions) == 0 {
		return nil
	}
	return e.admissions[len(e.admissions)-1].receipt
}

// settle drops the admissions whose receipts resolved, in admission order. A
// failed receipt fails every later admission through After, so settle drops
// them all, restores the replay save and the held head from before the first,
// and records their operations as failed. The caller holds the writer lock.
func (e *soEngine) settle() {
	for len(e.admissions) != 0 {
		// Stop at the first unresolved receipt.
		adm := e.admissions[0]
		select {
		case <-adm.receipt.Done():
		default:
			return
		}

		// A successful receipt makes its publication durable.
		if adm.receipt.Err() == nil {
			e.admissions = e.admissions[1:]
			continue
		}

		// A failed receipt takes its successors with it.
		for _, failed := range e.admissions {
			if failed.localID != "" {
				e.failed[failed.localID] = struct{}{}
			}
		}
		e.replay.unsave(adm.base, adm.span, adm.spanLen)
		e.retained = adm.retained
		e.admissions = nil
	}
}

// settled returns a channel that closes when the oldest unsettled receipt
// resolves, or nil when every admission settled. The caller holds the writer
// lock.
func (e *soEngine) settled() <-chan struct{} {
	if len(e.admissions) == 0 {
		return nil
	}
	return e.admissions[0].receipt.Done()
}

// pending reports whether an admission waits for its receipt. Direct writes of
// the replay save, the held roots and the SharedObject state wait until none
// does, since each would name blocks only the pending publications hold. The
// caller holds the writer lock.
func (e *soEngine) pending() bool {
	return len(e.admissions) != 0
}

// admittedState settles the resolved receipts and returns the admitted state,
// waiting for a snapshot without the operations of failed publications. The
// caller holds the writer lock.
func (e *soEngine) admittedState(ctx context.Context) (sobject.SharedObjectStateSnapshot, error) {
	// Settle first, so a failure just resolved is waited out.
	e.settle()
	ctr, rel, err := accessAdmittedState(ctx, e.so, nil)
	if err != nil {
		return nil, err
	}
	defer rel()

	// Wait for a current snapshot.
	return ctr.WaitValueWithValidator(ctx, func(snap sobject.SharedObjectStateSnapshot) (bool, error) {
		if snap == nil {
			return false, nil
		}
		stale, err := e.stale(ctx, snap)
		return !stale, err
	}, nil)
}

// stale reports whether snap holds an operation of a failed publication. The
// SharedObject publishes each snapshot after the state it shows, so a snapshot
// of the state before a failed receipt reset it can arrive after the reset.
// Every operation has a new local id, so a failed one never returns. The
// caller holds the writer lock.
func (e *soEngine) stale(ctx context.Context, snap sobject.SharedObjectStateSnapshot) (bool, error) {
	// Skip reading the set while nothing failed.
	if len(e.failed) == 0 {
		return false, nil
	}

	// Look for each failed operation.
	set, err := snap.GetOperationSet(ctx)
	if err != nil {
		return false, err
	}
	self := e.so.GetPeerID().String()
	for localID := range e.failed {
		if set.Find(self, localID) != nil {
			return true, nil
		}
	}
	return false, nil
}
