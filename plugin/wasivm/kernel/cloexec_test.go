package kernel

import "testing"

// TestCloexec checks the close-on-exec flag through pipe2, fcntl and dup, and
// that exec closes the flagged descriptors and keeps the others.
func TestCloexec(t *testing.T) {
	// Open a pipe with close-on-exec on both ends.
	p, mem := fsTestProcess()
	const pipefdPtr = 0x100
	if res := p.sysPipe2(pipefdPtr, guestOCLOEXEC); errnoFailed(res) {
		t.Fatalf("pipe2: %d", int64(res))
	}
	rfd, _ := mem.ReadUint32Le(pipefdPtr)
	wfd, _ := mem.ReadUint32Le(pipefdPtr + 4)
	if got := p.sysFcntl(rfd, fGETFD, 0); got != fdCLOEXEC {
		t.Fatalf("F_GETFD of the read end = %d, want FD_CLOEXEC", got)
	}

	// F_SETFD clears the flag, dup leaves it clear on the copy, and
	// F_DUPFD_CLOEXEC sets it.
	if res := p.sysFcntl(wfd, fSETFD, 0); res != 0 {
		t.Fatalf("F_SETFD: %d", int64(res))
	}
	dupFD := uint32(p.sysDup(rfd))
	cloexecDupFD := uint32(p.sysFcntl(wfd, fDUPFD_CLOEXEC, 10))
	for fd, want := range map[uint32]uint64{wfd: 0, dupFD: 0, cloexecDupFD: fdCLOEXEC} {
		if got := p.sysFcntl(fd, fGETFD, 0); got != want {
			t.Errorf("F_GETFD of %d = %d, want %d", fd, got, want)
		}
	}

	// Exec keeps only the descriptors without the flag.
	p.execChild([]string{"prog"}, nil, "/bin/prog")
	child := p.selfExecProc
	for _, fd := range []uint32{rfd, cloexecDupFD} {
		if _, ok := child.fds[fd]; ok {
			t.Errorf("exec kept close-on-exec descriptor %d", fd)
		}
	}
	for _, fd := range []uint32{wfd, dupFD} {
		if _, ok := child.fds[fd]; !ok {
			t.Errorf("exec closed descriptor %d", fd)
		}
	}
	if len(child.cloexec) != 0 {
		t.Errorf("exec left close-on-exec flags %v", child.cloexec)
	}
}
