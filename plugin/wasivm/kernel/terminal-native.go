//go:build !js

package kernel

import "golang.org/x/term"

// initTermOps sets platform-specific terminal operations for native builds.
func initTermOps(k *Kernel) {
	k.termMakeRaw = func(fd int) (any, error) {
		return term.MakeRaw(fd)
	}
	k.termRestore = func(fd int, state any) error {
		return term.Restore(fd, state.(*term.State))
	}
	k.termGetSize = func(fd int) (int, int, error) {
		return term.GetSize(fd)
	}
}
