//go:build windows

package launcher_install

import (
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// waitForProcessesUnder waits until no other process runs an executable under
// any of roots, such as the daemon's copy in the state root and Electron. It
// fails when one is still running at the deadline.
func waitForProcessesUnder(roots []string, timeout time.Duration) error {
	// Snapshot the running processes.
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return errors.Wrap(err, "list processes")
	}
	defer windows.CloseHandle(snapshot)

	// Wait on each process whose executable lies under a root.
	deadline := time.Now().Add(timeout)
	self := windows.GetCurrentProcessId()
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snapshot, &entry); err == nil; err = windows.Process32Next(snapshot, &entry) {
		if entry.ProcessID == self {
			continue
		}
		if err := waitForProcessUnder(entry.ProcessID, roots, deadline); err != nil {
			return err
		}
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return errors.Wrap(err, "list processes")
	}
	return nil
}

// waitForProcessUnder waits until deadline for process pid to exit when its
// executable lies under one of roots. Processes this user cannot open are
// not Spacewave's.
func waitForProcessUnder(pid uint32, roots []string, deadline time.Time) error {
	// Open the process to read its image path and wait on it.
	process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(process)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(process, 0, &buf[0], &size); err != nil {
		return nil
	}
	image := windows.UTF16ToString(buf[:size])
	if !underAny(image, roots) {
		return nil
	}

	// Wait for the exit event until the deadline.
	wait := max(time.Until(deadline), 0)
	event, err := windows.WaitForSingleObject(process, uint32(wait.Milliseconds()))
	if err != nil {
		return errors.Wrapf(err, "wait for %s", filepath.Base(image))
	}
	if event != windows.WAIT_OBJECT_0 {
		return errors.Errorf("%s is still running; close Spacewave, then uninstall again", filepath.Base(image))
	}
	return nil
}

// underAny reports whether path lies inside one of roots. Windows paths
// ignore case.
func underAny(path string, roots []string) bool {
	for _, root := range roots {
		prefix := strings.ToLower(filepath.Clean(root)) + `\`
		if strings.HasPrefix(strings.ToLower(path), prefix) {
			return true
		}
	}
	return false
}
