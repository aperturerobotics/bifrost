package kernel

import (
	"strconv"

	"github.com/pkg/errors"
)

var (
	// ErrNilConfig is returned by New for a nil configuration.
	ErrNilConfig = errors.New("config is nil")
	// ErrNoInit is returned by Boot when the init program is missing or
	// empty.
	ErrNoInit = errors.New("init program not found")
	// ErrNoStart is returned for a program that exports no _start function.
	ErrNoStart = errors.New("program has no _start export")
	// ErrReboot is returned by Boot when the system restarted through the
	// reboot call. The caller boots a new kernel.
	ErrReboot = errors.New("the system restarted")

	// errHalt ends a run that the reboot call halted or powered off.
	errHalt = errors.New("the system halted")

	// errNoProcessContext is returned by a host function called outside a
	// process.
	errNoProcessContext = errors.New("host call outside a process")
)

// ProcessExitError is a process exit. RunProcess returns it for a nonzero
// exit code, and exit panics with it to unwind the guest.
type ProcessExitError struct {
	// Code is the exit status.
	Code int
}

// Error implements error.
func (e *ProcessExitError) Error() string {
	return "process exited with code " + strconv.Itoa(e.Code)
}
