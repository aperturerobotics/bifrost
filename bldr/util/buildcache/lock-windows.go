//go:build !js && windows

package bldr_util_buildcache

import (
	"golang.org/x/sys/windows"
)

func (lock *Lock) lock() error {
	var overlapped windows.Overlapped
	return windows.LockFileEx(
		windows.Handle(lock.file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1,
		0,
		&overlapped,
	)
}

func (lock *Lock) unlock() error {
	var overlapped windows.Overlapped
	return windows.UnlockFileEx(
		windows.Handle(lock.file.Fd()),
		0,
		1,
		0,
		&overlapped,
	)
}
