package provider_local_test

import (
	"context"
	"slices"
	"testing"

	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/s4wave/spacewave/core/provider"
	provider_local "github.com/s4wave/spacewave/core/provider/local"
	"github.com/s4wave/spacewave/core/sobject"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/testbed"
)

// offlineMember is one member's mounted Space and the World it replays.
type offlineMember struct {
	t       *testing.T
	name    string
	tb      *testbed.Testbed
	account *provider_local.ProviderAccount
	object  *provider_local.SharedObject
	ref     *sobject.SharedObjectRef
	engine  sobject_world_engine.Engine
	release directive.Reference
}

// joinWriter invites a writer to the owner's Space over the owner's session
// transport and joins it as member. Returns the member's mounted Space and its
// reference.
func joinWriter(
	ctx context.Context,
	t *testing.T,
	ownerAccount *provider_local.ProviderAccount,
	ownerSession *provider_local.Session,
	ownerObject *provider_local.SharedObject,
	memberAccount *provider_local.ProviderAccount,
	memberSession *provider_local.Session,
) (*provider_local.SharedObject, *sobject.SharedObjectRef) {
	// Invite a writer over the owner's session transport.
	t.Helper()
	invite, err := ownerObject.CreateSOInviteOp(ctx, ownerObject.GetPrivKey(), "local", &sobject.SOInvite{Role: sobject.SOParticipantRole_SOParticipantRole_WRITER, MaxUses: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := ownerAccount.PrepareDirectInvite(ctx, ownerSession.GetPrivKey(), ownerObject.GetPrivKey(), invite); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ownerAccount.StopSessionTransport)
	t.Cleanup(ownerAccount.StopP2PSync)

	// Start the member's session transport and connect it to the owner's.
	if err := memberAccount.EnsureConfiguredSessionTransport(ctx, memberSession.GetPrivKey()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(memberAccount.StopSessionTransport)
	t.Cleanup(memberAccount.StopP2PSync)
	prepareSessionTransportBackends(ctx, t, ownerAccount.GetSessionTransport(), memberAccount.GetSessionTransport())

	// The member joins and mounts the Space.
	if _, err := memberAccount.JoinViaInvite(ctx, memberSession.GetPrivKey(), invite, ""); err != nil {
		t.Fatal(err)
	}
	memberObject := mountJoined(ctx, t, memberAccount, ownerObject.GetSharedObjectID())
	entries := memberAccount.GetSOListCtr().GetValue().GetSharedObjects()
	memberRef := entries[slices.IndexFunc(entries, func(entry *sobject.SharedObjectListEntry) bool {
		return entry.GetRef().GetProviderResourceRef().GetId() == ownerObject.GetSharedObjectID()
	})].GetRef()
	return memberObject, memberRef
}

// addLocalSession creates another local account and session on the provider of
// tb, as a second user of one daemon.
func addLocalSession(ctx context.Context, t *testing.T, tb *testbed.Testbed) (*provider_local.ProviderAccount, *provider_local.Session) {
	// Look up the local provider.
	t.Helper()
	prov, provRef, err := provider.ExLookupProvider(ctx, tb.Bus, "local", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(provRef.Release)
	localProv := prov.(*provider_local.Provider)

	// Create the account and its session.
	sessRef, err := localProv.CreateLocalAccountAndSession(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	// Access the session's account.
	accountID := sessRef.GetProviderResourceRef().GetProviderAccountId()
	acc, releaseAcc, err := localProv.AccessProviderAccount(ctx, accountID, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(releaseAcc)
	account := acc.(*provider_local.ProviderAccount)

	// Mount the session on it.
	sess, releaseSess, err := account.MountSession(ctx, sessRef, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(releaseSess)
	return account, sess.(*provider_local.Session)
}

// start starts the World engine and waits for its World.
func (m *offlineMember) start(ctx context.Context) {
	// Start the engine and wait for the replayed World.
	t := m.t
	t.Helper()
	conf := sobject_world_engine.NewConfig("offline-world-"+m.name, m.ref)
	eng, ref, err := world.StartEngine(ctx, m.tb.Bus, conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.release = ref
	m.engine = eng
}

// stop stops the World engine.
func (m *offlineMember) stop() {
	if m.release != nil {
		m.release.Release()
		m.release = nil
	}
}

// create creates object key in its own operation.
func (m *offlineMember) create(ctx context.Context, key string) {
	m.t.Helper()
	m.write(ctx, func(tx world.Tx) error {
		obj, _, err := world.CreateWorldObject(ctx, tx, key, func(cursor *block.Cursor) error {
			cursor.SetBlock(block_mock.NewExample(m.name+" wrote "+key), true)
			return nil
		})
		world.ReleaseObjectState(obj)
		return err
	})
}

// write commits one transaction.
func (m *offlineMember) write(ctx context.Context, fn func(tx world.Tx) error) {
	// Open the transaction.
	t := m.t
	t.Helper()
	tx, err := m.engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Apply and commit it.
	if err := fn(tx); err != nil {
		t.Fatalf("%s: %v", m.name, err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("%s commit: %v", m.name, err)
	}
}

// requireObjects checks that the World holds every key.
func (m *offlineMember) requireObjects(ctx context.Context, keys ...string) {
	// Read the World as of the latest operation set.
	t := m.t
	t.Helper()
	m.sync(ctx)
	tx, err := m.engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Look up each key.
	for _, key := range keys {
		obj, found, err := tx.GetObject(ctx, key)
		world.ReleaseObjectState(obj)
		if err != nil {
			t.Fatal(err)
		}
		if !found {
			t.Fatalf("%s World lacks %q", m.name, key)
		}
	}
}

// sync replays the latest operation set into the World, as each write
// transaction does before it forks.
func (m *offlineMember) sync(ctx context.Context) {
	m.t.Helper()
	tx, err := m.engine.NewTransaction(ctx, true)
	if err != nil {
		m.t.Fatal(err)
	}
	tx.Discard()
}

// root replays the latest operation set and returns the World root block.
func (m *offlineMember) root(ctx context.Context) *block.BlockRef {
	// Bring the World up to the latest operation set.
	t := m.t
	t.Helper()
	m.sync(ctx)

	// Read its root.
	var root *block.BlockRef
	if err := m.engine.AccessWorldState(ctx, nil, func(cursor *bucket_lookup.Cursor) error {
		root = cursor.GetRef().GetRootRef().CloneVT()
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return root
}

// health returns the current health of the Space.
func (m *offlineMember) health(ctx context.Context) *sobject.SharedObjectHealth {
	// Read the current health value.
	t := m.t
	t.Helper()
	healths, release, err := m.object.AccessSharedObjectHealth(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	return healths.GetValue()
}
