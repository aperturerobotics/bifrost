package s4wave_apt

import (
	"context"
	"errors"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	world_types "github.com/s4wave/spacewave/db/world/types"
)

func TestAptWorldOpsCreateTypedEntitiesAndRepositoryEdges(t *testing.T) {
	// Prepare a World and keys for the repository and its child records.
	ctx, ws := setupAptWorld(t)
	repositoryKey := "apt/repos/stable"
	packageKey := "apt/repos/stable/packages/busybox"
	buildSpecKey := "apt/repos/stable/specs/busybox"

	// Create the repository in the World.
	createRepo := NewCreateAptRepositoryOp(repositoryKey, &AptRepository{
		State:         AptRepositoryState_AptRepositoryState_EMPTY,
		Distribution:  "stable",
		Components:    []string{"main"},
		Architectures: []string{"i386"},
	})
	if _, _, err := ws.ApplyWorldOp(ctx, createRepo, ""); err != nil {
		t.Fatalf("ApplyWorldOp(create repository): %v", err)
	}

	// Import the package archive and verify the built state.
	aptPackage, debRef, err := ImportDebPackage(ctx, ws, repositoryKey, packageKey, testBusyboxDeb(t))
	if err != nil {
		t.Fatalf("ImportDebPackage: %v", err)
	}
	if got := aptPackage.GetState(); got != AptPackageState_AptPackageState_BUILT {
		t.Fatalf("imported package state = %s, want BUILT", got.String())
	}

	// Add a build specification referencing the source archive.
	sourceRef, err := block.BuildBlockRef([]byte("busybox-source"), nil)
	if err != nil {
		t.Fatalf("BuildBlockRef(source): %v", err)
	}
	addBuildSpec := NewAddAptBuildSpecOp(repositoryKey, buildSpecKey, &AptBuildSpec{
		SourcePackage: "busybox",
		SourceRef:     &bucket.ObjectRef{RootRef: sourceRef},
		Architectures: []string{"i386"},
	})
	if _, _, err := ws.ApplyWorldOp(ctx, addBuildSpec, ""); err != nil {
		t.Fatalf("ApplyWorldOp(add build spec): %v", err)
	}

	// Verify every created object has its expected type.
	assertObjectType(t, ctx, ws, repositoryKey, AptRepositoryTypeID)
	assertObjectType(t, ctx, ws, packageKey, AptPackageTypeID)
	assertObjectType(t, ctx, ws, buildSpecKey, AptBuildSpecTypeID)

	// Read back the repository, package, and build specification records.
	repository := readAptBlock[AptRepository](t, ctx, ws, repositoryKey)
	if repository.GetDistribution() != "stable" {
		t.Fatalf("repository distribution = %q, want stable", repository.GetDistribution())
	}
	storedPackage := readAptBlock[AptPackage](t, ctx, ws, packageKey)
	if storedPackage.GetName() != "busybox" {
		t.Fatalf("package name = %q, want busybox", storedPackage.GetName())
	}
	if !storedPackage.GetDebRef().EqualsRef(debRef) {
		t.Fatalf("package deb_ref = %s, want %s", storedPackage.GetDebRef().MarshalString(), debRef.MarshalString())
	}
	buildSpec := readAptBlock[AptBuildSpec](t, ctx, ws, buildSpecKey)
	if buildSpec.GetSourcePackage() != "busybox" {
		t.Fatalf("build spec source package = %q, want busybox", buildSpec.GetSourcePackage())
	}

	// Verify the repository graph links both child records.
	assertGraphEdge(t, ctx, ws, repositoryKey, PredAptRepoPackage.String(), packageKey)
	assertGraphEdge(t, ctx, ws, repositoryKey, PredAptRepoBuildSpec.String(), buildSpecKey)
}

func TestAptWorldOpsRejectInvalidInitialStates(t *testing.T) {
	// Verify a ready repository cannot be created as an initial record.
	indexRef, err := block.BuildBlockRef([]byte("dists-index"), nil)
	if err != nil {
		t.Fatalf("BuildBlockRef(index): %v", err)
	}
	createRepo := NewCreateAptRepositoryOp("apt/repos/stable", &AptRepository{
		State:         AptRepositoryState_AptRepositoryState_READY,
		Distribution:  "stable",
		Components:    []string{"main"},
		Architectures: []string{"i386"},
		IndexRef:      &bucket.ObjectRef{RootRef: indexRef},
	})
	if err := createRepo.Validate(); !errors.Is(err, ErrInvalidAptRepositoryInitialState) {
		t.Fatalf("CreateAptRepositoryOp.Validate err = %v, want invalid initial state", err)
	}

	// Verify built, published, and superseded packages cannot be added as initial records.
	debRef, err := block.BuildBlockRef([]byte("deb-payload"), nil)
	if err != nil {
		t.Fatalf("BuildBlockRef(deb): %v", err)
	}
	addBuiltPackage := NewAddAptPackageOp("apt/repos/stable", "apt/repos/stable/packages/busybox", &AptPackage{
		State:        AptPackageState_AptPackageState_BUILT,
		Name:         "busybox",
		Version:      "1:1.36.1-7",
		Architecture: "i386",
		DebRef:       debRef,
	})
	if err := addBuiltPackage.Validate(); !errors.Is(err, ErrInvalidAptPackageInitialState) {
		t.Fatalf("AddAptPackageOp.Validate built err = %v, want invalid initial state", err)
	}
	addPublishedPackage := NewAddAptPackageOp("apt/repos/stable", "apt/repos/stable/packages/busybox-published", &AptPackage{
		State:        AptPackageState_AptPackageState_PUBLISHED,
		Name:         "busybox",
		Version:      "1:1.36.1-7",
		Architecture: "i386",
		DebRef:       debRef,
	})
	if err := addPublishedPackage.Validate(); !errors.Is(err, ErrInvalidAptPackageInitialState) {
		t.Fatalf("AddAptPackageOp.Validate err = %v, want invalid initial state", err)
	}
	addSuperseded := NewAddAptPackageOp("apt/repos/stable", "apt/repos/stable/packages/busybox-old", &AptPackage{
		State:        AptPackageState_AptPackageState_SUPERSEDED,
		Name:         "busybox",
		Version:      "1:1.36.1-6",
		Architecture: "i386",
		DebRef:       debRef,
	})
	if err := addSuperseded.Validate(); !errors.Is(err, ErrInvalidAptPackageInitialState) {
		t.Fatalf("AddAptPackageOp.Validate superseded err = %v, want invalid initial state", err)
	}
}

func TestAptPublishPackageOpTransitionsBuiltPackage(t *testing.T) {
	// Prepare and publish a built World package.
	ctx, ws, packageKey := setupAptWorldWithBuiltPackage(t)
	publish := NewAptPublishPackageOp(packageKey)
	if _, _, err := ws.ApplyWorldOp(ctx, publish, ""); err != nil {
		t.Fatalf("ApplyWorldOp(publish): %v", err)
	}

	// Read back the package and verify its published state.
	aptPackage := readAptBlock[AptPackage](t, ctx, ws, packageKey)
	if got := aptPackage.GetState(); got != AptPackageState_AptPackageState_PUBLISHED {
		t.Fatalf("package state = %s, want PUBLISHED", got.String())
	}
}

func TestAptPublishPackageOpRejectsInvalidSourceState(t *testing.T) {
	// Create an importing World package for the rejected publish operation.
	ctx, ws := setupAptWorldWithRepository(t)
	packageKey := "apt/repos/stable/packages/busybox"
	addImporting := NewAddAptPackageOp("apt/repos/stable", packageKey, &AptPackage{
		State:        AptPackageState_AptPackageState_IMPORTING,
		Name:         "busybox",
		Version:      "1:1.36.1-7",
		Architecture: "i386",
		Description:  "Tiny utilities",
	})
	if _, _, err := ws.ApplyWorldOp(ctx, addImporting, ""); err != nil {
		t.Fatalf("ApplyWorldOp(add importing): %v", err)
	}

	// Verify publication rejects the importing package.
	publish := NewAptPublishPackageOp(packageKey)
	if _, _, err := ws.ApplyWorldOp(ctx, publish, ""); !errors.Is(err, ErrInvalidAptPackageStateTransition) {
		t.Fatalf("ApplyWorldOp(publish importing) err = %v, want invalid transition", err)
	}

	// Read back the package and verify its importing state.
	aptPackage := readAptBlock[AptPackage](t, ctx, ws, packageKey)
	if got := aptPackage.GetState(); got != AptPackageState_AptPackageState_IMPORTING {
		t.Fatalf("package state = %s, want IMPORTING", got.String())
	}
}

func TestAptSupersedePackageOpTransitionsPublishedPackage(t *testing.T) {
	// Prepare and publish the package before superseding it.
	ctx, ws, packageKey := setupAptWorldWithBuiltPackage(t)
	if _, _, err := ws.ApplyWorldOp(ctx, NewAptPublishPackageOp(packageKey), ""); err != nil {
		t.Fatalf("ApplyWorldOp(publish): %v", err)
	}

	// Supersede the published package in the World.
	if _, _, err := ws.ApplyWorldOp(ctx, NewAptSupersedePackageOp(packageKey), ""); err != nil {
		t.Fatalf("ApplyWorldOp(supersede): %v", err)
	}

	// Read back the package and verify its superseded state.
	aptPackage := readAptBlock[AptPackage](t, ctx, ws, packageKey)
	if got := aptPackage.GetState(); got != AptPackageState_AptPackageState_SUPERSEDED {
		t.Fatalf("package state = %s, want SUPERSEDED", got.String())
	}
}

func TestAptSupersedePackageOpRejectsInvalidSourceState(t *testing.T) {
	// Verify a built package cannot be superseded.
	ctx, ws, packageKey := setupAptWorldWithBuiltPackage(t)
	if _, _, err := ws.ApplyWorldOp(ctx, NewAptSupersedePackageOp(packageKey), ""); !errors.Is(err, ErrInvalidAptPackageStateTransition) {
		t.Fatalf("ApplyWorldOp(supersede built) err = %v, want invalid transition", err)
	}

	// Read back the package and verify its built state.
	aptPackage := readAptBlock[AptPackage](t, ctx, ws, packageKey)
	if got := aptPackage.GetState(); got != AptPackageState_AptPackageState_BUILT {
		t.Fatalf("package state = %s, want BUILT", got.String())
	}
}

func setupAptWorld(t *testing.T) (context.Context, world.WorldState) {
	// Start a World testbed and register its cleanup.
	t.Helper()
	ctx := t.Context()
	tb, err := world_testbed.Default(ctx, world_testbed.WithWorldVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Register the Apt operation controller on the testbed bus.
	opController := world.NewLookupOpController("test-apt-ops", tb.EngineID, world.NewSchema(RepositoryTypes).LookupOp)
	if _, err := tb.Bus.AddController(ctx, opController, nil); err != nil {
		t.Fatal(err)
	}

	return ctx, world.NewEngineWorldState(tb.Engine, true)
}

func setupAptWorldWithBuiltPackage(t *testing.T) (context.Context, world.WorldState, string) {
	// Import a package into the prepared World repository.
	t.Helper()
	ctx, ws := setupAptWorldWithRepository(t)
	packageKey := "apt/repos/stable/packages/busybox"
	if _, _, err := ImportDebPackage(ctx, ws, "apt/repos/stable", packageKey, testBusyboxDeb(t)); err != nil {
		t.Fatalf("ImportDebPackage: %v", err)
	}
	return ctx, ws, packageKey
}

func assertObjectType(t *testing.T, ctx context.Context, ws world.WorldState, objectKey, typeID string) {
	t.Helper()

	got, err := world_types.GetObjectType(ctx, ws, objectKey)
	if err != nil {
		t.Fatalf("GetObjectType(%s): %v", objectKey, err)
	}
	if got != typeID {
		t.Fatalf("object %s type = %q, want %q", objectKey, got, typeID)
	}
}

func readAptBlock[T any, PT block.Ptr[T]](
	t *testing.T,
	ctx context.Context,
	ws world.WorldState,
	objectKey string,
) PT {
	t.Helper()
	out, err := world.LookupObjectBody[T, PT](ctx, ws, objectKey)
	if err != nil {
		t.Fatalf("LookupObjectBody(%s): %v", objectKey, err)
	}
	return out
}

func assertGraphEdge(t *testing.T, ctx context.Context, ws world.WorldState, subjectKey, predicate, objectKey string) {
	t.Helper()

	quads, err := ws.LookupGraphQuads(ctx, world.NewGraphQuadWithKeys(subjectKey, predicate, objectKey, ""), 1)
	if err != nil {
		t.Fatalf("LookupGraphQuads(%s, %s, %s): %v", subjectKey, predicate, objectKey, err)
	}
	if len(quads) != 1 {
		t.Fatalf("expected graph edge %s --%s--> %s, got %d quads", subjectKey, predicate, objectKey, len(quads))
	}
}
