package resource_session

import (
	"context"
	"slices"

	"github.com/pkg/errors"
	account_settings "github.com/s4wave/spacewave/core/account/settings"
	resource_space "github.com/s4wave/spacewave/core/resource/space"
	"github.com/s4wave/spacewave/core/sobject"
	s4wave_session "github.com/s4wave/spacewave/sdk/session"
)

// SetAccountPlugin adds a plugin to or removes it from the account plugin
// list. Every Space the account has open follows the change.
func (r *SessionResource) SetAccountPlugin(
	ctx context.Context,
	req *s4wave_session.SetAccountPluginRequest,
) (*s4wave_session.SetAccountPluginResponse, error) {
	// Mount the account settings.
	so, release, err := r.mountAccountSettings(ctx)
	if err != nil {
		return nil, err
	}
	defer release()

	// Write the add or remove operation.
	op := &account_settings.AccountPluginOp{PluginId: req.GetPluginId()}
	settingsOp := &account_settings.AccountSettingsOp{}
	if req.GetEnabled() {
		settingsOp.Op = &account_settings.AccountSettingsOp_AddPlugin{AddPlugin: op}
	} else {
		settingsOp.Op = &account_settings.AccountSettingsOp_RemovePlugin{RemovePlugin: op}
	}
	opData, err := settingsOp.MarshalVT()
	if err != nil {
		return nil, errors.Wrap(err, "marshal account settings op")
	}
	if _, err := sobject.WriteOperation(ctx, so, opData, account_settings.ProcessAccountSettingsOps); err != nil {
		return nil, err
	}
	return &s4wave_session.SetAccountPluginResponse{}, nil
}

// watchAccountPlugins calls report with the account plugin list and again on
// each change until ctx ends. It is the AccountPluginSource of every Space the
// session mounts.
func (r *SessionResource) watchAccountPlugins(ctx context.Context, report func(ids []string)) error {
	// Mount the account settings and access their state.
	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()
	so, release, err := r.mountAccountSettings(ctx)
	if err != nil {
		return err
	}
	defer release()
	stateCtr, relStateCtr, err := so.AccessSharedObjectState(ctx, ctxCancel)
	if err != nil {
		return err
	}
	defer relStateCtr()

	// Report the list whenever it changes.
	var snap sobject.SharedObjectStateSnapshot
	var prev []string
	for first := true; ; first = false {
		snap, err = stateCtr.WaitValueChange(ctx, snap, nil)
		if err != nil {
			return err
		}
		settings, err := account_settings.ReadSnapshot(ctx, snap)
		if err != nil {
			return err
		}
		ids := settings.GetPluginIds()
		if !first && slices.Equal(prev, ids) {
			continue
		}
		report(ids)
		prev = ids
	}
}

// mountAccountSettings mounts the account settings SharedObject of the
// session's provider account. Call release when done.
func (r *SessionResource) mountAccountSettings(ctx context.Context) (sobject.SharedObject, func(), error) {
	// Resolve the account settings reference of the provider account.
	providerAcc := r.session.GetProviderAccount()
	settingsAcc, ok := providerAcc.(accountSettingsProvider)
	if !ok {
		return nil, nil, errors.New("provider account has no account settings")
	}
	settingsRef, err := settingsAcc.GetAccountSettingsRef(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Mount the settings SharedObject through the provider feature.
	soFeature, err := sobject.GetSharedObjectProviderAccountFeature(ctx, providerAcc)
	if err != nil {
		return nil, nil, err
	}
	so, release, err := soFeature.MountSharedObject(ctx, settingsRef, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "mount account settings")
	}
	return so, release, nil
}

// accountPluginSource returns the AccountPluginSource of the session's Spaces,
// or nil when its provider account keeps no account settings.
func (r *SessionResource) accountPluginSource() resource_space.AccountPluginSource {
	if _, ok := r.session.GetProviderAccount().(accountSettingsProvider); !ok {
		return nil
	}
	return r.watchAccountPlugins
}
