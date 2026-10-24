//go:build !js && !unix && !windows

package bldr_util_buildcache

import "sync"

var unsupportedFileLock sync.Mutex

func (lock *Lock) lock() error {
	unsupportedFileLock.Lock()
	return nil
}

func (lock *Lock) unlock() error {
	unsupportedFileLock.Unlock()
	return nil
}
