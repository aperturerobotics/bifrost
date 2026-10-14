package provider_local

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/kvtx"
	store_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// chainPublisher admits each state head to a publication the test resolves,
// as the publication writer resolves them: in admission order, applying each
// head to the store it publishes into.
type chainPublisher struct {
	// store is the object store the heads replace the state in.
	store kvtx.Store
	// heads are the admitted heads in admission order.
	heads []*block.AtomicHeadUpdate
	// receipts are the receipts of heads.
	receipts []*block.PublicationReceipt
	// onPublish, when set, runs after head i is recorded and before publish
	// returns, as the publisher does when a predecessor fails while a write
	// admits.
	onPublish func(i int)
}

// publish admits head and returns its pending receipt.
func (p *chainPublisher) publish(_ context.Context, head *block.AtomicHeadUpdate) (*block.PublicationReceipt, error) {
	receipt := block.NewPublicationReceipt()
	p.heads = append(p.heads, head)
	p.receipts = append(p.receipts, receipt)
	if p.onPublish != nil {
		p.onPublish(len(p.heads) - 1)
	}
	return receipt, nil
}

// apply writes head i to the store and resolves its receipt with the result.
func (p *chainPublisher) apply(ctx context.Context, i int) error {
	head := p.heads[i]
	err := kvtx.RunTransaction(ctx, true,
		func(ctx context.Context) (kvtx.Tx, error) {
			return p.store.NewTransaction(ctx, true)
		},
		func(ctx context.Context, tx kvtx.Tx) error {
			current, found, err := tx.Get(ctx, head.Key)
			if err != nil {
				return err
			}
			next, err := head.Replace(ctx, current, found)
			if err != nil {
				return err
			}
			return tx.Set(ctx, head.Key, next)
		},
	)
	p.receipts[i].Resolve(err)
	return err
}

// admissionHarness hosts a genesis shared object whose ordinary local
// operations publish through a chainPublisher.
type admissionHarness struct {
	// t is the running test.
	t *testing.T
	// store counts the direct state writes.
	store *commitCountStore
	// host adds the operations.
	host *sobject.SOHost
	// priv signs the operations.
	priv crypto.PrivKey
	// peerID is the peer of priv.
	peerID string
	// durable watches the durable state.
	durable ccontainer.Watchable[*sobject.SOState]
	// admitted watches the admitted state.
	admitted ccontainer.Watchable[*sobject.SOState]
	// pub admits the published state heads.
	pub *chainPublisher
}

// newAdmissionHarness seeds a genesis store and hosts it, publishing as
// "objs".
func newAdmissionHarness(t *testing.T) *admissionHarness {
	// Operation local ids are ULIDs, which reject the bubble's year 2000
	// clock, so move it forward.
	t.Helper()
	time.Sleep(time.Until(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))

	// Host a genesis store that publishes into its inner store.
	ctx := t.Context()
	inner := store_inmem.NewStore()
	store := &commitCountStore{Store: inner}
	_, _, priv := seedTestGenesisState(t, store)
	store.commits = 0
	watch, lock, syncFuncs := NewObjectStoreSOStateFuncs(ctx, store, "objs", "")
	host := sobject.NewSOHost(ctx, watch, lock, testSharedObjectID, syncFuncs)
	t.Cleanup(host.ClearContext)
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	// Watch both states of the entry.
	durable, relDurable, err := watch(ctx, testSharedObjectID, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(relDurable)
	admitted, relAdmitted, err := syncFuncs.Admitted(ctx, testSharedObjectID, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(relAdmitted)
	return &admissionHarness{
		t:        t,
		store:    store,
		host:     host,
		priv:     priv,
		peerID:   pid.String(),
		durable:  durable,
		admitted: admitted,
		pub:      &chainPublisher{store: inner},
	}
}

// add adds one operation, publishing it when publish is set, and returns its
// local id.
func (h *admissionHarness) add(publish bool) string {
	h.t.Helper()
	localID, err := h.addErr(publish)
	if err != nil {
		h.t.Fatal(err)
	}
	return localID
}

// addErr adds one operation, publishing it when publish is set.
func (h *admissionHarness) addErr(publish bool) (string, error) {
	ctx := h.t.Context()
	if publish {
		ctx = sobject.WithPublishState(ctx, h.pub.publish)
	}
	le := logrus.NewEntry(logrus.New())
	localID, _, err := h.host.AddLocalOperation(ctx, le, testStepFactorySet(), h.priv, []byte("op"))
	return localID, err
}

// expect checks which of the operations the state of w holds.
func (h *admissionHarness) expect(name string, w ccontainer.Watchable[*sobject.SOState], held map[string]bool) {
	h.t.Helper()
	state := w.GetValue()
	for localID, want := range held {
		op, err := state.GetOperation(h.peerID, localID)
		if err != nil {
			h.t.Fatal(err)
		}
		if (op != nil) != want {
			h.t.Fatalf("%s state holds %s = %v, want %v", name, localID, op != nil, want)
		}
	}
}

// TestSOStateAdmissionChain checks that a second operation admits and builds
// on the first while the first receipt is pending, and that the durable state
// shows each operation only after its receipt succeeds.
func TestSOStateAdmissionChain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Admit two operations without resolving either receipt.
		h := newAdmissionHarness(t)
		op1 := h.add(true)
		op2 := h.add(true)
		if len(h.pub.heads) != 2 {
			t.Fatalf("admitted %d heads, want 2", len(h.pub.heads))
		}
		h.expect("admitted", h.admitted, map[string]bool{op1: true, op2: true})
		h.expect("durable", h.durable, map[string]bool{op1: false, op2: false})
		if h.store.commits != 0 {
			t.Fatalf("admission committed %d direct writes", h.store.commits)
		}

		// Each receipt makes its operation durable, in order.
		ctx := t.Context()
		if err := h.pub.apply(ctx, 0); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		h.expect("durable", h.durable, map[string]bool{op1: true, op2: false})
		if err := h.pub.apply(ctx, 1); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		h.expect("durable", h.durable, map[string]bool{op1: true, op2: true})
		if h.durable.GetValue() != h.admitted.GetValue() {
			t.Fatal("settled durable state is not the admitted state")
		}
	})
}

// TestSOStateAdmissionFailure checks that a failed receipt fails the
// admission built on it and resets the admitted state to the durable state,
// which the next operation then builds on.
func TestSOStateAdmissionFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Admit two operations and fail the first.
		h := newAdmissionHarness(t)
		op1 := h.add(true)
		op2 := h.add(true)
		h.pub.receipts[0].Resolve(errors.New("publication failed"))
		synctest.Wait()
		held := map[string]bool{op1: false, op2: false}
		h.expect("admitted", h.admitted, held)
		h.expect("durable", h.durable, held)

		// The successor expects the failed state, so it fails too.
		ctx := t.Context()
		if err := h.pub.apply(ctx, 1); err == nil {
			t.Fatal("successor of a failed publication succeeded")
		}
		synctest.Wait()

		// The next operation builds on the durable state.
		op3 := h.add(true)
		if err := h.pub.apply(ctx, 2); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		held[op3] = true
		h.expect("admitted", h.admitted, held)
		h.expect("durable", h.durable, held)
	})
}

// TestSOStateAdmissionFailureDuringAdmit checks that an operation built on
// the admitted state before a receipt failed, and admitted after the entry
// reset that state, fails with the predecessor and leaves the admitted state
// at the durable state.
func TestSOStateAdmissionFailureDuringAdmit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Fail the first receipt while the second operation admits.
		h := newAdmissionHarness(t)
		op1 := h.add(true)
		h.pub.onPublish = func(i int) {
			if i != 1 {
				return
			}
			h.pub.receipts[0].Resolve(errors.New("publication failed"))
			synctest.Wait()
		}
		if _, err := h.addErr(true); !errors.Is(err, block.ErrPublicationDependency) {
			t.Fatalf("admission built on a failed publication returned %v, want ErrPublicationDependency", err)
		}

		// The failed chain leaves the durable state as the admitted state.
		h.expect("admitted", h.admitted, map[string]bool{op1: false})
		if h.durable.GetValue() != h.admitted.GetValue() {
			t.Fatal("admission built on a failed publication left the admitted state ahead")
		}

		// The dependent admission fails with its predecessor, and the next
		// operation builds on the durable state.
		ctx := t.Context()
		if err := h.pub.apply(ctx, 1); err == nil {
			t.Fatal("successor of a failed publication succeeded")
		}
		h.pub.onPublish = nil
		op3 := h.add(true)
		if err := h.pub.apply(ctx, 2); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()
		held := map[string]bool{op1: false, op3: true}
		h.expect("admitted", h.admitted, held)
		h.expect("durable", h.durable, held)
	})
}

// TestSOStateDirectWriteAfterChain checks that a direct state write started
// while a receipt is pending writes after it and includes its operation.
func TestSOStateDirectWriteAfterChain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		// Admit one operation and start a direct write behind it.
		h := newAdmissionHarness(t)
		op1 := h.add(true)
		var op2 string
		var addErr error
		done := make(chan struct{})
		go func() {
			defer close(done)
			op2, addErr = h.addErr(false)
		}()

		// The direct write waits for the receipt.
		synctest.Wait()
		select {
		case <-done:
			t.Fatal("direct write finished before the pending receipt")
		default:
		}
		if h.store.commits != 0 {
			t.Fatalf("direct write committed %d times before the receipt", h.store.commits)
		}

		// After the receipt it writes the state that includes the operation.
		if err := h.pub.apply(t.Context(), 0); err != nil {
			t.Fatal(err)
		}
		<-done
		if addErr != nil {
			t.Fatal(addErr)
		}
		if h.store.commits != 1 {
			t.Fatalf("direct write committed %d times, want 1", h.store.commits)
		}
		synctest.Wait()
		h.expect("durable", h.durable, map[string]bool{op1: true, op2: true})
		if h.durable.GetValue() != h.admitted.GetValue() {
			t.Fatal("direct write left the admitted state ahead")
		}
	})
}
