package kernel

// Linux errno values. wasm32 guests share the asm-generic numbering.
const (
	// EPERM reports an operation not permitted.
	EPERM = 1

	// ENOENT reports a missing file or directory.
	ENOENT = 2

	// ESRCH reports a missing process.
	ESRCH = 3

	// EINTR reports a call interrupted by a signal.
	EINTR = 4

	// EIO reports an input or output failure.
	EIO = 5

	// ENOEXEC reports a file that is neither a program nor a script.
	ENOEXEC = 8

	// EBADF reports a descriptor that is not open.
	EBADF = 9

	// ECHILD reports that there is no child to wait for.
	ECHILD = 10

	// EAGAIN reports a resource that is not ready yet.
	EAGAIN = 11

	// ENOMEM reports exhausted memory.
	ENOMEM = 12

	// EACCES reports denied permission.
	EACCES = 13

	// EFAULT reports a bad guest address.
	EFAULT = 14

	// EBUSY reports a resource in use, such as a mounted directory.
	EBUSY = 16

	// EEXIST reports a file that already exists.
	EEXIST = 17

	// ENODEV reports a filesystem type the kernel lacks.
	ENODEV = 19

	// ENOTDIR reports a path component that is not a directory.
	ENOTDIR = 20

	// EISDIR reports a directory where a file is required.
	EISDIR = 21

	// EINVAL reports an invalid argument.
	EINVAL = 22

	// ENOTTY reports a descriptor that is not a terminal.
	ENOTTY = 25

	// ESPIPE reports a seek on a pipe.
	ESPIPE = 29

	// EMLINK reports too many links.
	EMLINK = 31

	// EPIPE reports a write to a pipe with no reader.
	EPIPE = 32

	// ERANGE reports a result that does not fit its buffer.
	ERANGE = 34

	// ENOSYS reports a function the kernel does not implement.
	ENOSYS = 38

	// ENOTEMPTY reports a directory that is not empty.
	ENOTEMPTY = 39

	// ELOOP reports too many levels of symbolic links.
	ELOOP = 40

	// ENODATA reports a missing extended attribute.
	ENODATA = 61

	// ENOTSOCK reports a socket call on a descriptor that is not a socket.
	ENOTSOCK = 88

	// EOPNOTSUPP reports an operation the file does not support.
	EOPNOTSUPP = 95

	// EAFNOSUPPORT reports an address family without sockets.
	EAFNOSUPPORT = 97
)

// enosys is the result of a syscall the kernel does not implement.
const enosys = ^uint64(ENOSYS - 1)

// negErrno encodes errno as a syscall result: -errno as a uint64.
func negErrno(errno uint32) uint64 {
	return ^uint64(errno - 1)
}

// errnoFailed reports whether a syscall result is a negative errno rather than
// a count or success.
func errnoFailed(res uint64) bool {
	return res > 1<<63
}

// linuxToWasiErrno maps Linux errno values to WASI ones. wasi-libc numbers
// errno differently from Linux, so a POSIX host function translates the
// value before storing it in the guest's errno.
var linuxToWasiErrno = [EAFNOSUPPORT + 1]uint32{
	EPERM:        63,
	ENOENT:       44,
	ESRCH:        71,
	EINTR:        27,
	EIO:          29,
	ENOEXEC:      45,
	EBADF:        8,
	ECHILD:       12,
	EAGAIN:       6,
	ENOMEM:       48,
	EACCES:       2,
	EFAULT:       21,
	EBUSY:        10,
	EEXIST:       20,
	ENODEV:       43,
	ENOTDIR:      54,
	EISDIR:       31,
	EINVAL:       28,
	ENOTTY:       59,
	ESPIPE:       70,
	EMLINK:       34,
	EPIPE:        64,
	ERANGE:       68,
	ENOSYS:       52,
	ENOTEMPTY:    55,
	ELOOP:        32,
	ENODATA:      61,
	ENOTSOCK:     57,
	EOPNOTSUPP:   58,
	EAFNOSUPPORT: 5,
}

// wasiErrno translates a Linux errno to WASI. Values without a mapping pass
// through unchanged.
func wasiErrno(errno uint32) uint32 {
	if int(errno) < len(linuxToWasiErrno) && linuxToWasiErrno[errno] != 0 {
		return linuxToWasiErrno[errno]
	}
	return errno
}

// linuxErrno translates a WASI errno to Linux, the inverse of wasiErrno.
func linuxErrno(errno uint32) uint32 {
	for linux, wasi := range linuxToWasiErrno {
		if wasi == errno && wasi != 0 {
			return uint32(linux)
		}
	}
	return errno
}

// wasiErrnoFromResult returns the WASI errno of a syscall result, or success
// when the result is not a negative errno.
func wasiErrnoFromResult(result uint64) uint32 {
	if !errnoFailed(result) {
		return wasiErrnoSuccess
	}
	return wasiErrno(uint32(-result))
}
