//go:build !js

package callpath

import (
	"regexp"
	"slices"
	"testing"
)

const samplePkg = "github.com/s4wave/spacewave/bldr/devtool/callpath/testdata/sample"

func TestFindThroughInterfaceToLock(t *testing.T) {
	// Search from Run to a mutex lock.
	res, err := Find(Options{
		Patterns: []string{"./testdata/sample"},
		From:     regexp.MustCompile(`sample\.Run$`),
		To:       regexp.MustCompile(`^\(\*sync\.Mutex\)\.Lock$`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// The interface call reaches the lock through AddOp only.
	want := []Path{{
		samplePkg + ".Run",
		"(" + samplePkg + ".AddOp).Apply",
		"(*" + samplePkg + ".Store).add",
		"(*sync.Mutex).Lock",
	}}
	if res.Roots != 1 || !slices.EqualFunc(res.Paths, want, slices.Equal) {
		t.Fatalf("roots %d paths %q", res.Roots, res.Paths)
	}
	if !slices.Equal(res.Modules, []string{"github.com/s4wave/spacewave"}) {
		t.Fatalf("modules %q", res.Modules)
	}
}

func TestFindEveryRoot(t *testing.T) {
	// Search from every Apply method to the counter read.
	res, err := Find(Options{
		Patterns: []string{"./testdata/sample"},
		From:     regexp.MustCompile(`\)\.Apply$`),
		To:       regexp.MustCompile(`\.Store\)\.read$`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Only ReadOp reaches it.
	want := []Path{{
		"(" + samplePkg + ".ReadOp).Apply",
		"(*" + samplePkg + ".Store).read",
	}}
	if !slices.EqualFunc(res.Paths, want, slices.Equal) {
		t.Fatalf("roots %d paths %q", res.Roots, res.Paths)
	}
}
