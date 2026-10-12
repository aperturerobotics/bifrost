package bldr_plugin

import "context"

// InitialCapabilityRegistrationState reports the terminal state of a plugin
// instance's startup capability-registration pass.
type InitialCapabilityRegistrationState uint8

const (
	// InitialCapabilityRegistrationPending indicates the startup pass is running.
	InitialCapabilityRegistrationPending InitialCapabilityRegistrationState = iota
	// InitialCapabilityRegistrationComplete indicates the startup pass completed.
	InitialCapabilityRegistrationComplete
	// InitialCapabilityRegistrationFailed indicates the plugin instance ended
	// before completing the startup pass.
	InitialCapabilityRegistrationFailed
)

// InitialCapabilityRegistrar is a startup controller that publishes
// capabilities after it starts running. The plugin completes its initial
// capability registration only after the first pass of each one settles.
type InitialCapabilityRegistrar interface {
	// WaitInitialCapabilities returns once the first registration pass has
	// settled, whether it published its capabilities or failed.
	WaitInitialCapabilities(ctx context.Context) error
}
