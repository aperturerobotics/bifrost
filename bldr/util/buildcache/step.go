//go:build !js

package bldr_util_buildcache

// Step is the record of one completed build step.
type Step struct {
	// Inputs are the absolute paths of the files the step read.
	Inputs []string
	// Outputs are the paths of the files the step wrote, relative to the
	// output root.
	Outputs []string
	// Value is the step's own encoding of its result.
	Value []byte
}
