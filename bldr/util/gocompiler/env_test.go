package gocompiler

import (
	"bytes"
	"strings"
	"testing"
)

// TestGoCompilerPreservesModuleProxy checks the effective child-process setting.
func TestGoCompilerPreservesModuleProxy(t *testing.T) {
	// Preserve an explicit offline policy when launching the Go compiler.
	t.Setenv("GOPROXY", "off")
	cmd := NewGoCompilerCmd(t.Context(), "go", "env", "GOPROXY")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Compiler defaults must not replace the caller's module transport.
	if got := strings.TrimSpace(out.String()); got != "off" {
		t.Fatalf("GOPROXY = %q, want off", got)
	}
}
