package sobject

import (
	"context"
	"slices"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/net/crypto"
)

// RemoveSOParticipants removes participant configs and grants in one signed
// configuration change. It returns the peer IDs that were present and removed.
func RemoveSOParticipants(
	ctx context.Context,
	host *SOHost,
	targetPeerIDs []string,
	signerPriv crypto.PrivKey,
	revInfo *SORevocationInfo,
) ([]string, error) {
	targets := make(map[string]struct{}, len(targetPeerIDs))
	for _, peerID := range targetPeerIDs {
		if peerID != "" {
			targets[peerID] = struct{}{}
		}
	}
	if len(targets) == 0 {
		return nil, nil
	}

	state, err := host.GetHostState(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get current SO state")
	}
	currentCfg := state.GetConfig()
	if currentCfg == nil {
		return nil, nil
	}

	var removed []string
	for _, participant := range currentCfg.GetParticipants() {
		if _, ok := targets[participant.GetPeerId()]; ok {
			removed = append(removed, participant.GetPeerId())
		}
	}
	if len(removed) == 0 {
		return nil, nil
	}

	nextCfg := currentCfg.CloneVT()
	nextCfg.Participants = slices.DeleteFunc(nextCfg.Participants, func(participant *SOParticipantConfig) bool {
		_, ok := targets[participant.GetPeerId()]
		return ok
	})
	entry, err := BuildSOConfigChange(currentCfg, nextCfg, SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_REMOVE_PARTICIPANT, signerPriv, revInfo)
	if err != nil {
		return nil, errors.Wrap(err, "build config change")
	}
	if err := host.ApplyConfigChange(ctx, entry, func(state *SOState) error {
		state.RootGrants = slices.DeleteFunc(state.RootGrants, func(grant *SOGrant) bool {
			_, ok := targets[grant.GetPeerId()]
			return ok
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return removed, nil
}

// RemoveSOParticipant removes one participant config and grant.
func RemoveSOParticipant(
	ctx context.Context,
	host *SOHost,
	targetPeerID string,
	signerPriv crypto.PrivKey,
	revInfo *SORevocationInfo,
) (bool, error) {
	removed, err := RemoveSOParticipants(ctx, host, []string{targetPeerID}, signerPriv, revInfo)
	return len(removed) != 0, err
}
