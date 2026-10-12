package s4wave_vm

import "github.com/pkg/errors"

// normalizeDesiredState maps a requested VM state to RUNNING or STOPPED.
// STARTING and STOPPING are accepted as aliases for RUNNING and STOPPED.
func normalizeDesiredState(state VmState) (VmState, error) {
	switch state {
	case VmState_VmState_STARTING:
		return VmState_VmState_RUNNING, nil
	case VmState_VmState_STOPPING:
		return VmState_VmState_STOPPED, nil
	case VmState_VmState_STOPPED, VmState_VmState_RUNNING:
		return state, nil
	default:
		return VmState_VmState_STOPPED, errors.Errorf("invalid desired vm state %s", state.String())
	}
}

// nextRunStates returns the desired and observed states that begin a new run
// generation toward the requested state, given the stored desired and
// observed states. It rejects a request that would change nothing: starting a
// VM that is already starting or running, or stopping one already stopped.
// Starting a VM whose run failed is a retry.
func nextRunStates(desired, observed, requested VmState) (VmState, VmState, error) {
	// Normalize the requested and stored desired states.
	target, err := normalizeDesiredState(requested)
	if err != nil {
		return 0, 0, err
	}
	current := VmState_VmState_STOPPED
	if normalized, err := normalizeDesiredState(desired); err == nil {
		current = normalized
	}

	// Reject a request the VM already satisfies or is working toward.
	if target == current &&
		(target != VmState_VmState_RUNNING || observed != VmState_VmState_ERROR) &&
		(target != VmState_VmState_STOPPED || observed == VmState_VmState_STOPPED) {
		return 0, 0, errors.Errorf("vm desired state already %s", target.String())
	}

	// Observe the transition until the runtime reports its result.
	switch {
	case target == VmState_VmState_RUNNING:
		return target, VmState_VmState_STARTING, nil
	case observed != VmState_VmState_STOPPED:
		return target, VmState_VmState_STOPPING, nil
	default:
		return target, VmState_VmState_STOPPED, nil
	}
}
