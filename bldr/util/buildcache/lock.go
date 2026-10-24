//go:build !js

package bldr_util_buildcache

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Lock is an exclusive lock on a file, held across processes until Close.
type Lock struct {
	file *os.File
}

// LockFile creates lockPath and its parent directories if needed and waits
// for an exclusive lock on it.
func LockFile(lockPath string) (*Lock, error) {
	// Create the lock directory and open the lock file.
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, errors.Wrap(err, "open lock file")
	}

	// Take the exclusive file lock, closing the file on failure.
	lock := &Lock{file: file}
	if err := lock.lock(); err != nil {
		_ = file.Close()
		return nil, errors.Wrap(err, "lock file")
	}
	return lock, nil
}

// Close releases the lock and closes its file.
func (lock *Lock) Close() error {
	// Skip closing an already closed or empty lock.
	if lock == nil || lock.file == nil {
		return nil
	}

	// Unlock the file and close it, reporting the first error.
	unlockErr := lock.unlock()
	closeErr := lock.file.Close()
	lock.file = nil
	if unlockErr != nil {
		return errors.Wrap(unlockErr, "unlock file")
	}
	if closeErr != nil {
		return errors.Wrap(closeErr, "close lock file")
	}
	return nil
}
