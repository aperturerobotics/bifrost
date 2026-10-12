package s4wave_vm

import "testing"

func TestNextRunStates(t *testing.T) {
	const (
		stopped  = VmState_VmState_STOPPED
		starting = VmState_VmState_STARTING
		running  = VmState_VmState_RUNNING
		stopping = VmState_VmState_STOPPING
		failed   = VmState_VmState_ERROR
	)
	tests := []struct {
		name                         string
		desired, observed, requested VmState
		wantDesired, wantObserved    VmState
		wantChanged, wantErr         bool
	}{
		{"start a stopped vm", stopped, stopped, running, running, starting, true, false},
		{"starting alias", stopped, stopped, starting, running, starting, true, false},
		{"retry a failed run", running, failed, running, running, starting, true, false},
		{"stop a running vm", running, running, stopped, stopped, stopping, true, false},
		{"stop a failed run", running, failed, stopping, stopped, stopping, true, false},
		{"stop while stopping", stopped, running, stopped, stopped, stopping, true, false},
		{"start a running vm", running, running, running, 0, 0, false, false},
		{"start a starting vm", running, starting, running, 0, 0, false, false},
		{"stop a stopped vm", stopped, stopped, stopped, 0, 0, false, false},
		{"invalid request", stopped, stopped, failed, 0, 0, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired, observed, changed, err := nextRunStates(tt.desired, tt.observed, tt.requested)
			if (err != nil) != tt.wantErr || changed != tt.wantChanged {
				t.Fatalf("changed = %t, err = %v; want %t, error %t", changed, err, tt.wantChanged, tt.wantErr)
			}
			if changed && (desired != tt.wantDesired || observed != tt.wantObserved) {
				t.Fatalf("states = %s, %s; want %s, %s", desired, observed, tt.wantDesired, tt.wantObserved)
			}
		})
	}
}
