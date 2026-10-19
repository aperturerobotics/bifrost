package prom_plugin

import "context"

// TargetLister lists the World object keys of the targets a Queryable reads.
// The Space's set of prom/target objects is the list; a lister is where that
// set comes from.
type TargetLister interface {
	// ListTargets returns the object keys of the targets to query. A key whose
	// object does not exist is skipped by the query.
	ListTargets(ctx context.Context) ([]string, error)
}

// StaticTargets is a TargetLister with a fixed list of object keys.
type StaticTargets []string

// ListTargets returns the fixed object keys.
func (t StaticTargets) ListTargets(context.Context) ([]string, error) {
	return t, nil
}

// _ is a type assertion
var _ TargetLister = StaticTargets(nil)
