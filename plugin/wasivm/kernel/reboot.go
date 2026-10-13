package kernel

// The reboot commands: the magic values Linux's reboot call takes.
const (
	rbAutoboot   = 0x01234567
	rbHaltSystem = 0xcdef0123
	rbPowerOff   = 0x4321fedc
	rbEnableCAD  = 0x89abcdef
	rbDisableCAD = 0
)

// sysReboot ends the run Boot started: RB_AUTOBOOT restarts the system, and
// RB_HALT_SYSTEM and RB_POWER_OFF stop it. The Ctrl-Alt-Del commands change
// nothing, since the console sends no such key. On success the caller exits
// with the run.
func (p *Process) sysReboot(cmd uint32) uint64 {
	// Only root may restart or stop the system, and only a booted one.
	if p.uids.eff != 0 || p.kernel.shutdown == nil {
		return negErrno(EPERM)
	}

	// Choose the run's end.
	var cause error
	switch cmd {
	case rbEnableCAD, rbDisableCAD:
		return 0
	case rbAutoboot:
		cause = ErrReboot
	case rbHaltSystem, rbPowerOff:
		cause = errHalt
	default:
		return negErrno(EINVAL)
	}

	// End the run and the caller with it.
	p.kernel.shutdown(cause)
	panic(&ProcessExitError{})
}
