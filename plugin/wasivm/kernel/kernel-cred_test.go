package kernel

import "testing"

// TestIDs checks the id setters keep the real, effective, saved and
// filesystem ids as Linux does.
func TestIDs(t *testing.T) {
	tests := []struct {
		name string
		set  func(*ids)
		want ids
	}{
		{name: "setuid", set: func(s *ids) { s.set(1000) }, want: ids{1000, 1000, 1000, 1000}},
		{name: "seteuid", set: func(s *ids) { s.setRes(idKeep, 1000, idKeep) }, want: ids{0, 1000, 0, 1000}},
		{name: "setresuid", set: func(s *ids) { s.setRes(1, 2, 3) }, want: ids{1, 2, 3, 2}},
		{name: "setreuid real", set: func(s *ids) { s.setRe(1000, idKeep) }, want: ids{1000, 0, 0, 0}},
		{name: "setreuid effective", set: func(s *ids) { s.setRe(idKeep, 1000) }, want: ids{0, 1000, 1000, 1000}},
	}
	for _, tc := range tests {
		var s ids
		tc.set(&s)
		if s != tc.want {
			t.Fatalf("%s: ids = %+v, want %+v", tc.name, s, tc.want)
		}
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
