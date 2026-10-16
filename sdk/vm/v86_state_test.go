package s4wave_vm

import (
	"testing"
)

func TestV86CleanIDs(t *testing.T) {
	// Prepare the expected VM type and operation identifiers.
	cases := []struct {
		name string
		got  string
		want string
	}{
		{name: "vm type", got: VmV86TypeID, want: "vm/v86"},
		{name: "v86 image type", got: V86ImageTypeID, want: "vm/image/v86"},
		{name: "create vm op", got: CreateVmV86OpTypeID, want: "vm/v86/create"},
		{name: "set config op", got: SetV86ConfigOpTypeID, want: "vm/v86/set-config"},
		{name: "set state op", got: SetV86StateOpTypeID, want: "vm/v86/set-state"},
		{name: "create image op", got: CreateV86ImageOpTypeID, want: "vm/image/v86/create"},
		{name: "set image metadata op", got: SetV86ImageMetadataOpTypeID, want: "vm/image/v86/set-metadata"},
	}

	// Verify each VM identifier matches its public value.
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Fatalf("%s ID = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}
