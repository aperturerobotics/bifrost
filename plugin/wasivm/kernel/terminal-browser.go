//go:build js

package kernel

// initTermOps is a no-op on browser builds. Terminal operations are
// not available in the browser environment.
func initTermOps(k *Kernel) {}
