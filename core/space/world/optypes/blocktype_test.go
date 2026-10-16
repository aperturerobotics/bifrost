package optypes

import (
	"reflect"
	"testing"

	space_world "github.com/s4wave/spacewave/core/space/world"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/byteslice"
	s4wave_sql_query "github.com/s4wave/spacewave/sdk/sql/query"
	s4wave_sql_workbench "github.com/s4wave/spacewave/sdk/sql/workbench"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	s4wave_wizard "github.com/s4wave/spacewave/sdk/world/wizard"
)

// TestLookupBlockType checks that the Space schema constructs the core block
// types by their stored IDs.
//
// VM images resolve in core because the CDN catalog Space that lists them is
// mounted without plugins. The SQL roots resolve only through the SQL plugin,
// which keeps the SQL packages out of the core GoScript build.
func TestLookupBlockType(t *testing.T) {
	for _, tc := range []struct {
		typeID string
		want   block.Block
	}{
		{space_world.SpaceSettingsTypeID, (*space_world.SpaceSettings)(nil)},
		{"github.com/s4wave/spacewave/core/space/world.SpaceSettings", (*space_world.SpaceSettings)(nil)},
		{s4wave_vm.V86ImageTypeID, (*s4wave_vm.V86Image)(nil)},
		{s4wave_vm.WasiVmImageTypeID, (*s4wave_vm.WasiVmImage)(nil)},
		{"sync/app/notes", (*byteslice.ByteSlice)(nil)},
		{"wizard/git", (*s4wave_wizard.WizardState)(nil)},
		{s4wave_sql_query.SqlQueryTypeID, nil},
		{s4wave_sql_workbench.SqlWorkbenchTypeID, nil},
	} {
		// Resolve the block type, then check its identity and constructor.
		got, err := LookupBlockType(t.Context(), tc.typeID)
		if err != nil {
			t.Fatalf("LookupBlockType(%s): %v", tc.typeID, err)
		}
		if tc.want == nil {
			if got != nil {
				t.Fatalf("LookupBlockType(%s) = %v, want nil", tc.typeID, got)
			}
			continue
		}
		if got == nil || got.GetBlockTypeID() != tc.typeID {
			t.Fatalf("LookupBlockType(%s) = %v, want its block type", tc.typeID, got)
		}
		if blk := got.Constructor(); reflect.TypeOf(blk) != reflect.TypeOf(tc.want) {
			t.Fatalf("LookupBlockType(%s) constructs %T, want %T", tc.typeID, blk, tc.want)
		}
	}
}
