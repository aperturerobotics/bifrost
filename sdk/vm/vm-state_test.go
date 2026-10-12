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
		wantErr                      bool
	}{
		{"start a stopped vm", stopped, stopped, running, running, starting, false},
		{"starting alias", stopped, stopped, starting, running, starting, false},
		{"retry a failed run", running, failed, running, running, starting, false},
		{"stop a running vm", running, running, stopped, stopped, stopping, false},
		{"stop a failed run", running, failed, stopping, stopped, stopping, false},
		{"stop while stopping", stopped, running, stopped, stopped, stopping, false},
		{"start a running vm", running, running, running, 0, 0, true},
		{"start a starting vm", running, starting, running, 0, 0, true},
		{"stop a stopped vm", stopped, stopped, stopped, 0, 0, true},
		{"invalid request", stopped, stopped, failed, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desired, observed, err := nextRunStates(tt.desired, tt.observed, tt.requested)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, want error %t", err, tt.wantErr)
			}
			if err == nil && (desired != tt.wantDesired || observed != tt.wantObserved) {
				t.Fatalf("states = %s, %s; want %s, %s", desired, observed, tt.wantDesired, tt.wantObserved)
			}
		})
	}
}
