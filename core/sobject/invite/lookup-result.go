package sobject_invite

import (
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/crypto"
)

// InviteLookupResult contains the resolved invite and its context.
type InviteLookupResult struct {
	// Host is the SOHost managing the shared object.
	Host *sobject.SOHost
	// InviteMutator applies invite state mutations for this shared object.
	InviteMutator sobject.InviteMutator
	// Invite is the matching SOInvite.
	Invite *sobject.SOInvite
	// SharedObjectID is the ID of the shared object.
	SharedObjectID string
	// OwnerPrivKey is the owner's private key for signing config changes.
	OwnerPrivKey crypto.PrivKey
}
