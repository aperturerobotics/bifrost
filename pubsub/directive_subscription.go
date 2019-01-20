package pubsub

import (
	"errors"
	"time"

	"github.com/aperturerobotics/controllerbus/directive"
)

// BuildChannelSubscription is a directive to subscribe to a channel.
type BuildChannelSubscription interface {
	// Directive indicates BuildChannelSubscription is a directive.
	directive.Directive

	// BuildChannelSubscriptionChannelID returns the channel ID constraint.
	// Cannot be empty.
	BuildChannelSubscriptionChannelID() string
}

// BuildChannelSubscriptionValue is the result type for BuildChannelSubscription.
// The value is removed and replaced when necessary.
type BuildChannelSubscriptionValue = Subscription

// buildChannelSubscription implements BuildChannelSubscription
type buildChannelSubscription struct {
	channelID string
}

// NewBuildChannelSubscription constructs a new BuildChannelSubscription directive.
func NewBuildChannelSubscription(channelID string) BuildChannelSubscription {
	return &buildChannelSubscription{channelID: channelID}
}

// Validate validates the directive.
// This is a cursory validation to see if the values "look correct."
func (d *buildChannelSubscription) Validate() error {
	if d.channelID == "" {
		return errors.New("channel id cannot be empty")
	}

	return nil
}

// GetValueBuildChannelSubscriptionOptions returns options relating to value handling.
func (d *buildChannelSubscription) GetValueOptions() directive.ValueOptions {
	return directive.ValueOptions{
		// UnrefDisposeDur is the duration to wait to dispose a directive after all
		// references have been released.
		UnrefDisposeDur: time.Second * 3,
	}
}

// BuildChannelSubscriptionChannelID returns the channel ID constraint.
func (d *buildChannelSubscription) BuildChannelSubscriptionChannelID() string {
	return d.channelID
}

// IsEquivalent checks if the other directive is equivalent. If two
// directives are equivalent, and the new directive does not superceed the
// old, then the new directive will be merged (de-duplicated) into the old.
func (d *buildChannelSubscription) IsEquivalent(other directive.Directive) bool {
	od, ok := other.(BuildChannelSubscription)
	if !ok {
		return false
	}

	return d.BuildChannelSubscriptionChannelID() == od.BuildChannelSubscriptionChannelID()
}

// Superceeds checks if the directive overrides another.
// The other directive will be canceled if superceded.
func (d *buildChannelSubscription) Superceeds(other directive.Directive) bool {
	return false
}

// GetName returns the directive's type name.
// This is not necessarily unique, and is primarily intended for display.
func (d *buildChannelSubscription) GetName() string {
	return "BuildChannelSubscription"
}

// GetDebugString returns the directive arguments stringified.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (d *buildChannelSubscription) GetDebugVals() directive.DebugValues {
	vals := directive.DebugValues{}
	vals["channel-id"] = []string{d.BuildChannelSubscriptionChannelID()}
	return vals
}

// _ is a type assertion
var _ BuildChannelSubscription = ((*buildChannelSubscription)(nil))
