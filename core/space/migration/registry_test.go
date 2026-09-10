package space_migration_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	space_migration "github.com/s4wave/spacewave/core/space/migration"
	objecttypes "github.com/s4wave/spacewave/core/space/world/objecttypes"
	volume_world "github.com/s4wave/spacewave/db/volume/world"
)

// TestBuiltInsAreClassifiedAndBoundToCentralInventory verifies complete migration coverage.
func TestBuiltInsAreClassifiedAndBoundToCentralInventory(t *testing.T) {
	// Compare the migration registry with the application inventory.
	registry, err := space_migration.BuiltInRegistry()
	if err != nil {
		t.Fatal(err)
	}
	central := objecttypes.BuiltInObjectTypeIDs()
	if !slices.Equal(central, registry.TypeIDs()) {
		t.Fatalf("migration registry diverges from central ObjectType inventory: central=%v migration=%v", central, registry.TypeIDs())
	}

	// Require a policy for each type and preserve installation identity boundaries.
	for _, typeID := range central {
		handler := registry.Lookup(typeID)
		if handler == nil {
			t.Fatalf("built-in type %q has no handler", typeID)
		}
		if handler.Classification() == space_migration.ClassificationUnclassified {
			t.Fatalf("built-in type %q is unclassified", typeID)
		}
		if typeID == volume_world.ObjectTypeID && handler.Classification() != space_migration.ClassificationNonMigratable {
			t.Fatal("Volume backing must not migrate installation identity as an ordinary object")
		}
	}
}

// TestSchemaHandlersRejectSyntheticMetadata requires the live source payload.
func TestSchemaHandlersRejectSyntheticMetadata(t *testing.T) {
	// Resolve handlers with real payload requirements.
	registry, err := space_migration.BuiltInRegistry()
	if err != nil {
		t.Fatal(err)
	}

	// Caller-provided references cannot substitute for a stored payload.
	for _, typeID := range []string{"spacewave/secret", "canvas"} {
		handler := registry.Lookup(typeID)
		if handler == nil {
			t.Fatalf("handler %q is missing", typeID)
		}
		_, err := handler.Rewrite(context.Background(), &space_migration.ObjectDescriptor{
			ObjectKey:  typeID,
			ObjectType: typeID,
			References: []space_migration.TypedReference{{Kind: space_migration.ReferenceExternal, Value: "synthetic"}},
		}, space_migration.NewIdentityMap())
		if !errors.Is(err, space_migration.ErrPayloadSchemaRefused) {
			t.Fatalf("handler %q accepted synthetic metadata: %v", typeID, err)
		}
	}
}
