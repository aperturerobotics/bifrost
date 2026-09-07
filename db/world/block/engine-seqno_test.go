package world_block

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/coord"
	coord_inmem "github.com/s4wave/spacewave/db/coord/inmem"
	"github.com/s4wave/spacewave/db/world"
	world_mock "github.com/s4wave/spacewave/db/world/mock"
)

// TestEngineGetSeqnoRefreshesDurableHead checks revision reads after another
// engine publishes a new durable root, with and without coordinated snapshots.
func TestEngineGetSeqnoRefreshesDurableHead(t *testing.T) {
	for _, coordinated := range []bool{false, true} {
		name := "shared-head"
		if coordinated {
			name = "coordinator"
		}
		t.Run(name, func(t *testing.T) {
			// Keep both engines on the same real store with independent heads.
			ctx := t.Context()
			writer := newRetirementTestEngine(t, ctx)
			var coordinator coord.Coordinator
			if coordinated {
				coordinator = coord_inmem.NewCoordinator()
			}
			root := writer.baseRoot.Clone()
			t.Cleanup(root.Release)
			reader, err := NewEngine(
				ctx,
				writer.le,
				root,
				world_mock.LookupMockOp,
				nil,
				false,
				WithWriteCoordinator(coordinator, coord.Scope{VolumeID: "seqno-volume", ObjectStoreID: "seqno-store"}, nil,
					func(context.Context) (*bucket.ObjectRef, error) { return writer.GetRootRef(), nil }),
			)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := reader.Close(); err != nil {
					t.Error(err)
				}
			})
			before, err := reader.GetSeqno(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// Publish an operation through the writer without touching the reader.
			ws := world.NewEngineWorldState(writer, true)
			_, _, err = world.CreateWorldObject(ctx, ws, "seqno/example", func(cursor *block.Cursor) error {
				cursor.SetBlock(block_mock.NewExample("before"), true)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			want, _, err := ws.ApplyWorldOp(ctx, world_mock.NewMockWorldOp("seqno/example", "after"), "")
			if err != nil {
				t.Fatal(err)
			}
			if want <= before {
				t.Fatalf("writer revision = %d, want greater than %d", want, before)
			}

			// A direct revision read must discover the newly published root.
			got, err := reader.GetSeqno(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("reader revision = %d, want %d", got, want)
			}
		})
	}
}
