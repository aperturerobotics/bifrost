//go:build !js && unix

package bldr_util_buildcache

import (
	stderrors "errors"

	"golang.org/x/sys/unix"
)

func (lock *Lock) lock() error {
	for {
		err := unix.Flock(int(lock.file.Fd()), unix.LOCK_EX)
		if stderrors.Is(err, unix.EINTR) {
			continue
		}
		return err
	}
}

func (lock *Lock) unlock() error {
	return unix.Flock(int(lock.file.Fd()), unix.LOCK_UN)
}
