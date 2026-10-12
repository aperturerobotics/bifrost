package kernel

import "testing"

// TestSetuidSetsAllUids checks setuid sets the real, effective, saved, and
// filesystem uid together.
func TestSetuidSetsAllUids(t *testing.T) {
	p, _ := newMemProcess(wasmPageSize)
	if res := p.sysSetuid(1000); res != 0 {
		t.Fatalf("setuid: got %#x, want 0", res)
	}
	if p.uid != 1000 || p.euid != 1000 || p.suid != 1000 || p.fsuid != 1000 {
		t.Fatalf("uids = %d/%d/%d/%d, want all 1000", p.uid, p.euid, p.suid, p.fsuid)
	}
}

// TestSetgidSetsAllGids checks setgid sets the real, effective, saved, and
// filesystem gid together.
func TestSetgidSetsAllGids(t *testing.T) {
	p, _ := newMemProcess(wasmPageSize)
	if res := p.sysSetgid(1000); res != 0 {
		t.Fatalf("setgid: got %#x, want 0", res)
	}
	if p.gid != 1000 || p.egid != 1000 || p.sgid != 1000 || p.fsgid != 1000 {
		t.Fatalf("gids = %d/%d/%d/%d, want all 1000", p.gid, p.egid, p.sgid, p.fsgid)
	}
}

// TestGetgroupsSizeTooSmall checks getgroups rejects a non-zero size below the
// group count with EINVAL.
func TestGetgroupsSizeTooSmall(t *testing.T) {
	p, _ := newMemProcess(wasmPageSize)
	p.groups = []uint32{1, 2, 3}
	if res := p.sysGetgroups(2, 64); res != negErrno(EINVAL) {
		t.Fatalf("getgroups(2): got %#x, want EINVAL %#x", res, negErrno(EINVAL))
	}
}
