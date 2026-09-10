package provider_local

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/util/ulid"
	"github.com/s4wave/spacewave/core/session"
	session_controller "github.com/s4wave/spacewave/core/session/controller"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/blocktype"
	blocktype_controller "github.com/s4wave/spacewave/db/blocktype/controller"
	"github.com/s4wave/spacewave/net/transport/inproc"
)

// TestAccountMergeReturningSession keeps a source replica offline during merge,
// then removes the merging machine before that replica discovers the redirect.
func TestAccountMergeReturningSession(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	t.Cleanup(cancel)
	network := inproc.NewNetwork()
	a, aSession, _ := setupMigrationClient(ctx, t, network)
	b, bSession, bController := setupMigrationClient(ctx, t, network)
	c, cSession, cController := setupMigrationClient(ctx, t, network)
	var spaces []*sobject.SharedObjectRef
	var leaves []*block.BlockRef
	var payloads [][]byte
	for _, account := range []*ProviderAccount{a, b} {
		ref, err := account.CreateSharedObject(ctx, ulid.NewULID(), &sobject.SharedObjectMeta{BodyType: "space"}, "", "")
		if err != nil {
			t.Fatal(err)
		}
		leaf, payload := seedAccountReplicaPayload(ctx, t, account, ref)
		spaces = append(spaces, ref)
		leaves = append(leaves, leaf)
		payloads = append(payloads, payload)
	}

	// C is an independent source-account Session with its own storage key.
	cSource, cSourceSession := enrollMeshReplica(ctx, t, b, bSession, c, cSession)
	cEntry, err := cController.RegisterSession(ctx, cSourceSession.GetSessionRef(), nil)
	if err != nil {
		t.Fatal(err)
	}
	waitReplicaCopy(ctx, t, cSource, spaces[1])
	cSource.StopP2PSync()
	cSource.StopSessionTransport()

	// The merging machine first receives the selected destination through the
	// same native enrollment used by the pairing engine.
	destination, destinationSession := enrollMeshReplica(ctx, t, a, aSession, b, bSession)
	if _, err := bController.RegisterSession(ctx, destinationSession.GetSessionRef(), nil); err != nil {
		t.Fatal(err)
	}
	commit, err := b.MergePairingAccount(ctx, bSession, destination, destinationSession.GetSessionRef())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := commit(ctx); err != nil {
		t.Fatal(err)
	}
	if active := destination.GetSessionTransport(); active == nil || active.GetPeerID() != bSession.GetPeerId() {
		t.Fatal("merged account still connects with its temporary enrollment key")
	}
	// A second move must preserve the destination known to still-offline C.
	if _, err := destination.MergePairingAccount(ctx, destinationSession, c, cSession.GetSessionRef()); err == nil || !strings.Contains(err.Error(), "previous account merge") {
		t.Fatalf("second merge did not retain the offline Session recovery destination: %v", err)
	}
	movedSpace := sobject.NewSharedObjectRef(a.GetProviderID(), a.GetAccountID(), spaces[1].GetProviderResourceRef().GetId(), spaces[1].GetBlockStoreId())
	waitReplicaObject(ctx, t, a, movedSpace)
	waitReplicaCopy(ctx, t, a, movedSpace)
	sourceSettings, err := b.GetAccountSettingsRef(ctx)
	if err != nil {
		t.Fatal(err)
	}
	recovery := sobject.NewSharedObjectRef(a.GetProviderID(), a.GetAccountID(), sourceSettings.GetProviderResourceRef().GetId(), sourceSettings.GetBlockStoreId())
	waitReplicaObject(ctx, t, a, recovery)
	b.StopP2PSync()
	b.StopSessionTransport()
	destination.StopP2PSync()
	destination.StopSessionTransport()

	// Only A can now provide the source's signed recovery state and destination.
	if err := cSource.EnsureConfiguredSessionTransport(ctx, cSourceSession.GetPrivKey()); err != nil {
		t.Fatal(err)
	}
	if err := cSource.StartPersistentP2PSync(ctx, cSource.GetSessionTransport()); err != nil {
		t.Fatal(err)
	}
	var returned *session.SessionListEntry
	for {
		var changed <-chan struct{}
		cController.GetSessionBroadcast().HoldLock(func(_ func(), getWait func() <-chan struct{}) { changed = getWait() })
		returned, err = cController.GetSessionByIdx(ctx, cEntry.GetSessionIndex())
		if err != nil {
			t.Fatal(err)
		}
		if returned.GetSessionRef().GetProviderResourceRef().GetProviderAccountId() == a.GetAccountID() {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatalf("returning Session did not follow the committed account: %v", returned)
		case <-changed:
		}
	}
	account, releaseAccount, err := c.t.p.AccessProviderAccount(ctx, a.GetAccountID(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(releaseAccount)
	replica := account.(*ProviderAccount)
	for _, original := range spaces {
		ref := sobject.NewSharedObjectRef(a.GetProviderID(), a.GetAccountID(), original.GetProviderResourceRef().GetId(), original.GetBlockStoreId())
		waitReplicaObject(ctx, t, replica, ref)
		waitReplicaCopy(ctx, t, replica, ref)
	}
	a.StopP2PSync()
	a.StopSessionTransport()
	for i, original := range spaces {
		ref := sobject.NewSharedObjectRef(a.GetProviderID(), a.GetAccountID(), original.GetProviderResourceRef().GetId(), original.GetBlockStoreId())
		object, releaseObject, err := replica.MountSharedObject(ctx, ref, nil)
		if err != nil {
			t.Fatal(err)
		}
		data, found, err := object.GetBlockStore().(*BlockStore).store.GetBlock(ctx, leaves[i])
		releaseObject()
		if err != nil || !found || !bytes.Equal(data, payloads[i]) {
			t.Fatalf("returning Session lacks an independent file copy: found=%v err=%v", found, err)
		}
	}
}

func setupMigrationClient(ctx context.Context, t *testing.T, network *inproc.Network) (*ProviderAccount, *Session, session.SessionController) {
	t.Helper()
	tb, _, account, mounted, release := setupProviderAndSessionInternal(ctx, t, func(p *Provider) {
		p.localNetwork = transport.WithInprocNetwork(network)
	})
	t.Cleanup(release)
	if err := account.EnsureConfiguredSessionTransport(ctx, mounted.GetPrivKey()); err != nil {
		t.Fatal(err)
	}
	decoder := blocktype_controller.NewController(func(_ context.Context, typeID string) (blocktype.BlockType, error) {
		if typeID == "test/replica-payload" {
			return blocktype.NewBlockType(typeID, block_mock.NewRootBlock), nil
		}
		return nil, nil
	})
	releaseDecoder, err := tb.Bus.AddController(ctx, decoder, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(releaseDecoder)
	tb.StaticResolver.AddFactory(session_controller.NewFactory(tb.Bus))
	_, reference, err := tb.Bus.AddDirective(resolver.NewLoadControllerWithConfig(&session_controller.Config{VolumeId: tb.EngineVolumeID}), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reference.Release)
	controller, lookup, err := session.ExLookupSessionController(ctx, tb.Bus, "", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lookup.Release)
	if _, err := controller.RegisterSession(ctx, mounted.GetSessionRef(), nil); err != nil {
		t.Fatal(err)
	}
	return account, mounted, controller
}
