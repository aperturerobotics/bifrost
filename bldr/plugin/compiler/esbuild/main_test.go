//go:build !js

package esbuild

import (
	"context"
	"os"
	"testing"

	starlark "github.com/s4wave/spacewave/bldr/project/starlark"
)

// TestMain handles the bounded evaluator command before starting fixture tests.
func TestMain(m *testing.M) {
	// Run the production evaluator protocol when validation starts this binary.
	if len(os.Args) == 2 && os.Args[1] == starlark.BoundedCommand {
		if err := starlark.RunBounded(context.Background(), os.Stdin, os.Stdout); err != nil {
			// A failed diagnostic write cannot change the child's unsuccessful exit.
			_, _ = os.Stderr.WriteString(err.Error() + "\n")
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Run the fixtures only in the parent process.
	os.Exit(m.Run())
}
