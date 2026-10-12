package kernel

import "testing"

// TestGetrlimitDefaults checks the finite NOFILE default and the unbounded
// default of a resource with no limit set.
func TestGetrlimitDefaults(t *testing.T) {
	// Read the NOFILE limit.
	p, mem := newMemProcess(wasmPageSize)
	if res := p.sysGetrlimit(rlimitNofile, 0); res != 0 {
		t.Fatalf("getrlimit(NOFILE) = %#x, want 0", res)
	}
	cur, _ := mem.ReadUint64Le(0)
	hard, _ := mem.ReadUint64Le(8)
	if cur != 1024 || hard != 4096 {
		t.Fatalf("NOFILE = %d/%d, want 1024/4096", cur, hard)
	}

	// Read the CPU limit, resource 0.
	if res := p.sysGetrlimit(0, 16); res != 0 {
		t.Fatalf("getrlimit(CPU) = %#x, want 0", res)
	}
	if cur, _ := mem.ReadUint64Le(16); cur != rlimInfinity {
		t.Fatalf("CPU soft limit = %#x, want infinity", cur)
	}
}

// TestSetrlimitRoundTrip checks getrlimit reads back what setrlimit stored.
func TestSetrlimitRoundTrip(t *testing.T) {
	// Set the NOFILE limit.
	p, mem := newMemProcess(wasmPageSize)
	mem.WriteUint64Le(0, 256)
	mem.WriteUint64Le(8, 512)
	if res := p.sysSetrlimit(rlimitNofile, 0); res != 0 {
		t.Fatalf("setrlimit = %#x, want 0", res)
	}

	// Read it back.
	if res := p.sysGetrlimit(rlimitNofile, 32); res != 0 {
		t.Fatalf("getrlimit = %#x, want 0", res)
	}
	cur, _ := mem.ReadUint64Le(32)
	hard, _ := mem.ReadUint64Le(40)
	if cur != 256 || hard != 512 {
		t.Fatalf("NOFILE = %d/%d, want 256/512", cur, hard)
	}
}

// TestSetrlimitRejectsSoftAboveHard checks setrlimit fails with EINVAL when
// the soft limit exceeds the hard limit.
func TestSetrlimitRejectsSoftAboveHard(t *testing.T) {
	p, mem := newMemProcess(wasmPageSize)
	mem.WriteUint64Le(0, 1000)
	mem.WriteUint64Le(8, 500)
	if res := p.sysSetrlimit(rlimitNofile, 0); res != negErrno(EINVAL) {
		t.Fatalf("setrlimit = %#x, want EINVAL", res)
	}
}

// TestRlimitRejectsBadResource checks getrlimit and setrlimit fail with EINVAL
// for an unknown resource.
func TestRlimitRejectsBadResource(t *testing.T) {
	p, _ := newMemProcess(wasmPageSize)
	if res := p.sysGetrlimit(rlimitNlimits, 0); res != negErrno(EINVAL) {
		t.Fatalf("getrlimit = %#x, want EINVAL", res)
	}
	if res := p.sysSetrlimit(rlimitNlimits, 0); res != negErrno(EINVAL) {
		t.Fatalf("setrlimit = %#x, want EINVAL", res)
	}
}
