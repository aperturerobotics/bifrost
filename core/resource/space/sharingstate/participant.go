package sharingstate

import (
	"slices"
	"strings"

	"github.com/s4wave/spacewave/core/sobject"
)

// ParticipantPresentation contains account labels used by sharing rows.
type ParticipantPresentation struct {
	// SelfAccountID is the viewer's account ID.
	SelfAccountID string
	// SelfEntityID is the viewer's entity label, shown for its own account.
	SelfEntityID string
	// AccountLabels maps account IDs to their entity labels.
	AccountLabels map[string]string
}

// ParticipantInfo is the app-facing participant presentation row.
type ParticipantInfo struct {
	// AccountID is the participant's account, empty for a bare peer.
	AccountID string
	// EntityID is the account's entity label.
	EntityID string
	// PeerIDs are the participant's peers.
	PeerIDs []string
	// Role is the highest role among the participant's peers.
	Role sobject.SOParticipantRole
	// IsSelf is set when one of the peers is the viewer.
	IsSelf bool
}

// ViewerRole returns the current viewer's effective participant role.
func ViewerRole(state *sobject.SOState, peerID string) sobject.SOParticipantRole {
	if state == nil || peerID == "" {
		return sobject.SOParticipantRole_SOParticipantRole_UNKNOWN
	}

	role := sobject.SOParticipantRole_SOParticipantRole_UNKNOWN
	for _, participant := range state.GetConfig().GetParticipants() {
		if participant.GetPeerId() != peerID {
			continue
		}
		if participant.GetRole() > role {
			role = participant.GetRole()
		}
	}
	return role
}

// BuildParticipantInfo groups participant rows by account or peer identity.
func BuildParticipantInfo(
	soState *sobject.SOState,
	selfPeerID string,
	presentation *ParticipantPresentation,
) []*ParticipantInfo {
	// Return no rows until the SharedObject configuration is available.
	if soState == nil || soState.GetConfig() == nil {
		return nil
	}

	// Read the configured audience before allocating participant rows.
	participants := soState.GetConfig().GetParticipants()
	if len(participants) == 0 {
		return nil
	}

	// Group each participant by account identity, falling back to its peer.
	rows := make(map[string]*ParticipantInfo, len(participants))
	keys := make([]string, 0, len(participants))
	for _, participant := range participants {
		peerID := participant.GetPeerId()
		if peerID == "" {
			continue
		}

		accountID := participant.GetEntityId()
		key := accountID
		if key == "" {
			key = "peer:" + peerID
		}

		row := rows[key]
		if row == nil {
			row = &ParticipantInfo{
				AccountID: accountID,
				Role:      participant.GetRole(),
			}
			if accountID != "" && presentation != nil {
				if label := presentation.AccountLabels[accountID]; label != "" {
					row.EntityID = label
				}
				if accountID == presentation.SelfAccountID && presentation.SelfEntityID != "" {
					row.EntityID = presentation.SelfEntityID
				}
			}
			rows[key] = row
			keys = append(keys, key)
		}

		if participant.GetRole() > row.Role {
			row.Role = participant.GetRole()
		}
		row.PeerIDs = append(row.PeerIDs, peerID)
		if peerID == selfPeerID {
			row.IsSelf = true
		}
	}

	// Return no rows when the audience has no usable peer identities.
	if len(keys) == 0 {
		return nil
	}

	// Sort participant rows by their visible label.
	slices.SortStableFunc(keys, func(a, b string) int {
		return strings.Compare(participantSortLabel(rows[a]), participantSortLabel(rows[b]))
	})

	// Materialize the ordered participant rows for the caller.
	out := make([]*ParticipantInfo, 0, len(keys))
	for _, key := range keys {
		out = append(out, rows[key])
	}
	return out
}

// participantSortLabel returns the label a participant row sorts by: its
// entity, account, or first peer.
func participantSortLabel(info *ParticipantInfo) string {
	if info == nil {
		return ""
	}
	if info.EntityID != "" {
		return info.EntityID
	}
	if info.AccountID != "" {
		return info.AccountID
	}
	if len(info.PeerIDs) != 0 {
		return info.PeerIDs[0]
	}
	return ""
}

// equalParticipantInfo reports whether two participant rows hold the same
// values.
func equalParticipantInfo(a *ParticipantInfo, b *ParticipantInfo) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.AccountID == b.AccountID &&
		a.EntityID == b.EntityID &&
		a.Role == b.Role &&
		a.IsSelf == b.IsSelf &&
		slices.Equal(a.PeerIDs, b.PeerIDs)
}
