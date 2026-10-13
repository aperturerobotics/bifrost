package kernel

import (
	"bytes"
	"testing"
)

// TestUname checks uname lays out the names as a wasm32 struct utsname and
// reports the name sethostname set.
func TestUname(t *testing.T) {
	// Name the host, then read the names back.
	p, mem := newMemProcess(wasmPageSize)
	const namePtr, utsPtr = 64, 256
	mem.Write(namePtr, []byte("box"))
	if res := p.sysSetUTSName(&p.kernel.hostname, namePtr, 3); res != 0 {
		t.Fatalf("sethostname = %#x", res)
	}
	if res := p.sysUname(utsPtr); res != 0 {
		t.Fatalf("uname = %#x", res)
	}

	// Check each field.
	for i, want := range []string{"Linux", "box", utsRelease, utsVersion, "wasm32", ""} {
		field, _ := mem.Read(utsPtr+uint32(i*utsFieldLen), utsFieldLen)
		if got := string(field[:bytes.IndexByte(field, 0)]); got != want {
			t.Fatalf("field %d = %q, want %q", i, got, want)
		}
	}

	// A name without room for its NUL is too long.
	if res := p.sysSetUTSName(&p.kernel.hostname, namePtr, utsFieldLen); res != negErrno(EINVAL) {
		t.Fatalf("sethostname of %d bytes = %#x, want EINVAL", utsFieldLen, res)
	}
}
