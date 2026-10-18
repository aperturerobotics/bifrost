package bldr_manifest_world

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// TestExStoreManifestOpSkipsStoredManifest checks that storing an already
// stored and linked manifest applies no World operation, and that a missing
// link still applies one.
func TestExStoreManifestOpSkipsStoredManifest(t *testing.T) {
	// Start a testbed.
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tb.Release()

	// Build the mock World state from an empty cursor.
	ocs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer ocs.Release()
	mockWS, err := world_block.BuildMockWorldState(ctx, le, true, ocs, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create two manifest stores and count operations applied to the World.
	const hostKey, otherKey = "plugin-host", "other-host"
	for _, key := range []string{hostKey, otherKey} {
		if _, err := CreateManifestStore(ctx, mockWS, key); err != nil {
			t.Fatal(err.Error())
		}
	}
	ws := &opCountingWorldState{WorldState: mockWS}

	// Store the manifest once, then store it again unchanged.
	ref := createTestManifestRef(t, ctx, tb, "spacewave-web", "js", 3)
	const manifestKey = "plugin-host/spacewave-web"
	for range 2 {
		if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, ref); err != nil {
			t.Fatal(err.Error())
		}
	}
	if ws.applied != 1 {
		t.Fatalf("applied %d operations storing one manifest twice, want 1", ws.applied)
	}

	// A new link applies an operation for the stored manifest.
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey, otherKey}, ref); err != nil {
		t.Fatal(err.Error())
	}
	if ws.applied != 2 {
		t.Fatalf("applied %d operations after adding a link, want 2", ws.applied)
	}
}

// TestExStoreManifestOpStoresReleaseAuthorization checks that an authorized
// ManifestRef is stored once, collected with its authorization, and stored
// again when only the authorization changes.
func TestExStoreManifestOpStoresReleaseAuthorization(t *testing.T) {
	// Start a testbed.
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tb.Release()

	// Build the mock World state from an empty cursor.
	ocs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer ocs.Release()
	mockWS, err := world_block.BuildMockWorldState(ctx, le, true, ocs, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create the manifest store and count operations applied to the World.
	const hostKey, manifestKey = "release", "release/spacewave-web"
	if _, err := CreateManifestStore(ctx, mockWS, hostKey); err != nil {
		t.Fatal(err.Error())
	}
	ws := &opCountingWorldState{WorldState: mockWS}

	// Sign the manifest reference with a fresh release key on each call.
	ref := createTestManifestRef(t, ctx, tb, "spacewave-web", "js", 3)
	signRef := func() {
		key, _, err := crypto.GenerateEd25519Key(nil)
		if err != nil {
			t.Fatal(err.Error())
		}
		if err := ref.SignReleaseAuthorization(key); err != nil {
			t.Fatal(err.Error())
		}
	}

	// Store the authorized manifest twice.
	signRef()
	for range 2 {
		if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, ref); err != nil {
			t.Fatal(err.Error())
		}
	}
	if ws.applied != 1 {
		t.Fatalf("applied %d operations storing one authorized manifest twice, want 1", ws.applied)
	}

	// Collect the manifest with the stored authorization.
	collected, errs, err := CollectManifests(ctx, ws, nil, hostKey)
	if err != nil || len(errs) != 0 {
		t.Fatalf("collect manifests: %v %v", err, errs)
	}
	got := collected["spacewave-web"]
	if len(got) != 1 || !got[0].ManifestRef.EqualVT(ref.GetManifestRef()) {
		t.Fatalf("collected %v, want the stored manifest", got)
	}
	if !got[0].ReleaseAuthorization.EqualVT(ref.GetReleaseAuthorization()) {
		t.Fatal("collected manifest lost its release authorization")
	}

	// A new authorization of the same manifest applies an operation.
	signRef()
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, ref); err != nil {
		t.Fatal(err.Error())
	}
	if ws.applied != 2 {
		t.Fatalf("applied %d operations after a new authorization, want 2", ws.applied)
	}
}

// newReleaseTestWorld builds a World with a manifest store at hostKey.
func newReleaseTestWorld(t *testing.T, hostKey string) (context.Context, *testbed.Testbed, world.WorldState) {
	// Start a testbed.
	t.Helper()
	ctx := t.Context()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(tb.Release)

	// Build the mock World state from an empty cursor.
	ocs, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Cleanup(ocs.Release)
	ws, err := world_block.BuildMockWorldState(ctx, le, true, ocs, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create the manifest store.
	if _, err := CreateManifestStore(ctx, ws, hostKey); err != nil {
		t.Fatal(err.Error())
	}
	return ctx, tb, ws
}

// TestExStoreManifestOpKeepsAuthorizationOfSameManifest checks that storing an
// unauthorized ref of a manifest onto its authorized ManifestRef object, as a
// download does when it copies the manifest into the World bucket, keeps the
// authorization and the ManifestRef object type.
func TestExStoreManifestOpKeepsAuthorizationOfSameManifest(t *testing.T) {
	// Build a World with a manifest store.
	const hostKey, manifestKey = "release", "release/spacewave-web"
	ctx, tb, ws := newReleaseTestWorld(t, hostKey)

	// Store the authorized manifest at an external bucket.
	ref := createTestManifestRef(t, ctx, tb, "spacewave-web", "js", 3)
	key, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := ref.SignReleaseAuthorization(key); err != nil {
		t.Fatal(err.Error())
	}
	external := ref.CloneVT()
	external.ManifestRef.BucketId = "external-bucket"
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, external); err != nil {
		t.Fatal(err.Error())
	}

	// Store the same manifest again without an authorization, in the World bucket.
	local := ref.CloneVT()
	local.ReleaseAuthorization = nil
	local.ManifestRef.BucketId = ""
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, local); err != nil {
		t.Fatal(err.Error())
	}

	// The object stays an authorized ManifestRef at the World bucket.
	if err := world_types.CheckObjectType(ctx, ws, manifestKey, ManifestRefTypeID); err != nil {
		t.Fatal(err.Error())
	}
	got, _, err := LookupManifestRef(ctx, ws, manifestKey)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !got.GetReleaseAuthorization().EqualVT(ref.GetReleaseAuthorization()) {
		t.Fatal("stored manifest ref lost its release authorization")
	}
	if got.GetManifestRef().GetBucketId() == "external-bucket" {
		t.Fatal("stored manifest ref was not relocated")
	}
	collected, errs, err := CollectManifests(ctx, ws, nil, hostKey)
	if err != nil || len(errs) != 0 {
		t.Fatalf("collect manifests: %v %v", err, errs)
	}
	if got := collected["spacewave-web"]; len(got) != 1 || got[0].ReleaseAuthorization == nil {
		t.Fatalf("collected %v, want the authorized manifest", got)
	}
}

// TestExStoreManifestOpRetypesReplacedManifestRef checks that storing a
// different unauthorized manifest onto an authorized ManifestRef object
// replaces both its root and its type.
func TestExStoreManifestOpRetypesReplacedManifestRef(t *testing.T) {
	// Build a World with a manifest store.
	const hostKey, manifestKey = "release", "release/spacewave-web"
	ctx, tb, ws := newReleaseTestWorld(t, hostKey)

	// Store an authorized manifest, then a different unauthorized one.
	authorized := createTestManifestRef(t, ctx, tb, "spacewave-web", "js", 3)
	key, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := authorized.SignReleaseAuthorization(key); err != nil {
		t.Fatal(err.Error())
	}
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, authorized); err != nil {
		t.Fatal(err.Error())
	}
	next := createTestManifestRef(t, ctx, tb, "spacewave-web", "js", 4)
	if err := ExStoreManifestOp(ctx, ws, peer.ID("test"), manifestKey, []string{hostKey}, next); err != nil {
		t.Fatal(err.Error())
	}

	// The object is a Manifest whose type and root agree.
	if err := world_types.CheckObjectType(ctx, ws, manifestKey, ManifestTypeID); err != nil {
		t.Fatal(err.Error())
	}
	m, _, err := LookupManifest(ctx, ws, manifestKey)
	if err != nil {
		t.Fatal(err.Error())
	}
	if m.GetMeta().GetRev() != 4 {
		t.Fatalf("stored manifest rev %d, want 4", m.GetMeta().GetRev())
	}
	collected, errs, err := CollectManifests(ctx, ws, nil, hostKey)
	if err != nil || len(errs) != 0 {
		t.Fatalf("collect manifests: %v %v", err, errs)
	}
	if got := collected["spacewave-web"]; len(got) != 1 || got[0].ReleaseAuthorization != nil {
		t.Fatalf("collected %v, want one unauthorized manifest", got)
	}
}

// opCountingWorldState counts the operations applied to a WorldState.
type opCountingWorldState struct {
	world.WorldState
	// applied is the number of ApplyWorldOp calls.
	applied int
}

// ApplyWorldOp counts op and applies it to the wrapped WorldState.
func (w *opCountingWorldState) ApplyWorldOp(ctx context.Context, op world.Operation, sender peer.ID) (uint64, bool, error) {
	w.applied++
	return w.WorldState.ApplyWorldOp(ctx, op, sender)
}
