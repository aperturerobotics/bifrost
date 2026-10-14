package binpkg

import (
	"slices"
	"strings"
	"testing"
)

// TestResolve checks that an index resolves to the runtime closure in
// dependency order, choosing the first present alternative of an any-of group
// and leaving out blockers and provided packages.
func TestResolve(t *testing.T) {
	// Read an index whose htop needs a provided package, an any-of group and a
	// blocker.
	const index = "ARCH: wasm32\n\n" +
		"CPV: sys-process/htop-3.4.1\nLICENSE: GPL-2+\n" +
		"RDEPEND: >=sys-libs/ncurses-6:0=[unicode(+)] sys-libs/musl || ( app-misc/absent sys-apps/hwdata ) !sys-process/htop-legacy\nPATH: sys-process/htop/htop-3.4.1-1.gpkg.tar\n\n" +
		"CPV: sys-libs/ncurses-6.5_p20250125-r1\nPATH: sys-libs/ncurses/ncurses-6.5_p20250125-r1-1.gpkg.tar\n\n" +
		"CPV: sys-apps/hwdata-0.390\n\n"
	pkgs, err := ReadIndex(strings.NewReader(index))
	if err != nil {
		t.Fatal(err)
	}
	if v := pkgs["sys-libs/ncurses"].Version; v != "6.5_p20250125-r1" {
		t.Fatalf("ncurses version %q", v)
	}

	// Resolve htop and compare the plan.
	plan, err := Resolve(pkgs, map[string]bool{"sys-libs/musl": true}, []string{"sys-process/htop"})
	if err != nil {
		t.Fatal(err)
	}
	var cpvs []string
	for _, pkg := range plan {
		cpvs = append(cpvs, pkg.CPV)
	}
	want := []string{"sys-libs/ncurses-6.5_p20250125-r1", "sys-apps/hwdata-0.390", "sys-process/htop-3.4.1"}
	if !slices.Equal(cpvs, want) {
		t.Fatalf("plan %v, want %v", cpvs, want)
	}
}
