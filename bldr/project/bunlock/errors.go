package bunlock

import "github.com/pkg/errors"

var (
	// ErrTrustedDependencies indicates that the lockfile permits install scripts.
	ErrTrustedDependencies = errors.New("bun.lock trusts dependencies to run their install scripts")
	// ErrIntegrity indicates that a downloaded tarball differs from its lock pin.
	ErrIntegrity = errors.New("npm tarball integrity mismatch")
)
