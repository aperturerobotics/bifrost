package sobject_sync

import (
	"context"
	"strings"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/peer"
)

// TestSnapshotExchangeRequiresHeldAuthority drives the real packet exchange
// with valid readable candidates that differ only at the authority boundary.
func TestSnapshotExchangeRequiresHeldAuthority(t *testing.T) {
	// Establish authority before accepting anything from the remote stream.
	const soID = "snapshot-held-authority"
	owner := mustKeyPair(t)
	reader := mustKeyPair(t)
	local := mustKeyPair(t)
	localID, err := peer.IDFromPrivateKey(local)
	if err != nil {
		t.Fatal(err)
	}
	initial := &sobject.SOState{
		Config: &sobject.SharedObjectConfig{Participants: []*sobject.SOParticipantConfig{
			participantCfg(mustPeerIDStr(t, owner), sobject.SOParticipantRole_SOParticipantRole_OWNER),
			participantCfg(mustPeerIDStr(t, reader), sobject.SOParticipantRole_SOParticipantRole_READER),
			participantCfg(localID.String(), sobject.SOParticipantRole_SOParticipantRole_WRITER),
		}},
		Root: &sobject.SORoot{InnerSeqno: 1},
	}
	trustSnapshotConfig(t, initial, owner)
	signSnapshotRoot(t, soID, initial, owner)
	initial.RootGrants = append(initial.RootGrants, buildGrant(t, soID, owner, local.GetPublic()))

	// Each rejection preserves every held byte, including configuration and queue state.
	for _, test := range []struct {
		// name identifies the attempted authority violation.
		name string
		// mutate changes an otherwise valid candidate and, when needed, its held checkpoint.
		mutate func(held, candidate *sobject.SOState)
		// wantError names the decisive boundary rather than an unrelated earlier failure.
		wantError string
		// advanceHeld commits a local configuration change during the access check.
		advanceHeld bool
	}{
		{name: "authorized sequence jump"},
		{name: "configuration advances during access check", advanceHeld: true, wantError: "configuration authority"},
		{name: "same head self promotion", mutate: func(_, candidate *sobject.SOState) {
			candidate.Config.Participants[1].Role = sobject.SOParticipantRole_SOParticipantRole_OWNER
			signSnapshotRoot(t, soID, candidate, reader)
		}, wantError: "configuration authority"},
		{name: "chainless different head", mutate: func(_, candidate *sobject.SOState) {
			candidate.Config.ConfigChainHash[0] ^= 1
			candidate.Config.ConfigChainSeqno++
		}, wantError: "configuration authority"},
		{name: "empty held checkpoint", mutate: func(held, candidate *sobject.SOState) {
			held.Config.ConfigChainHash = nil
			candidate.Config = held.Config.CloneVT()
		}, wantError: "configuration authority"},
		{name: "reader signed root", mutate: func(_, candidate *sobject.SOState) {
			signSnapshotRoot(t, soID, candidate, reader)
		}, wantError: "root authority"},
		{name: "tampered root", mutate: func(_, candidate *sobject.SOState) {
			candidate.Root.Inner[0] ^= 1
		}, wantError: "root authority"},
		{name: "unsigned root", mutate: func(_, candidate *sobject.SOState) {
			candidate.Root.ValidatorSignatures = nil
		}, wantError: "consensus"},
		{name: "duplicate signer", mutate: func(_, candidate *sobject.SOState) {
			candidate.Root.ValidatorSignatures = append(candidate.Root.ValidatorSignatures, candidate.Root.ValidatorSignatures[0].CloneVT())
		}, wantError: "root authority"},
	} {
		t.Run(test.name, func(t *testing.T) {
			held := initial.CloneVT()
			candidate := held.CloneVT()
			candidate.Root.InnerSeqno = 5
			signSnapshotRoot(t, soID, candidate, owner)
			if test.mutate != nil {
				test.mutate(held, candidate)
			}
			before := held.CloneVT()
			host, ctr := newMemHost(soID, held)
			t.Cleanup(host.ClearContext)
			var accessChecks []SnapshotAccessValidator
			if test.advanceHeld {
				accessChecks = append(accessChecks, func(ctx context.Context, _ *sobject.SOState) error {
					entry, err := sobject.BuildSOConfigChange(held.GetConfig(), held.GetConfig(), sobject.SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_ADD_INVITE, owner, nil)
					if err != nil {
						return err
					}
					if err := host.ApplyConfigChange(ctx, entry, nil); err != nil {
						return err
					}
					before = ctr.GetValue().CloneVT()
					return nil
				})
			}
			syncer := NewSOSync(gateLogger(), nil, soID, localID, local, host, accessChecks...)
			data, err := candidate.MarshalVT()
			if err != nil {
				t.Fatal(err)
			}
			snapshot := &SOSyncSnapshot{SoState: data, RootSeqno: 5}
			if len(held.GetConfig().GetConfigChainHash()) == 0 {
				err = syncer.applyPeerSnapshot(t.Context(), gateLogger(), snapshot)
			} else {
				err = runSnapshotExchange(t, syncer, t.Context(), &SOSyncMessage{
					Body: &SOSyncMessage_Snapshot{Snapshot: snapshot},
				})
			}
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
				if !ctr.GetValue().EqualVT(candidate) {
					t.Fatal("authorized sequence jump did not converge")
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want %q", err, test.wantError)
			}
			if !ctr.GetValue().EqualVT(before) {
				t.Fatal("rejected snapshot changed held state")
			}
		})
	}
}
