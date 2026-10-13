package sobject

import "slices"

// Clone copies the state while sharing its immutable signed operations.
// The operation slice and every other field are independent. Callers must not
// modify a shared operation or its signature; CloneVT makes a fully deep copy.
func (s *SOState) Clone() *SOState {
	// Preserve nil snapshots.
	if s == nil {
		return nil
	}

	// Deep-copy the remaining fields without copying operation bytes.
	rest := *s
	rest.Ops = nil
	next := rest.CloneVT()
	next.Ops = slices.Clone(s.Ops)
	return next
}
