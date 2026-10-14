package sharingstate

// MailboxEntry contains mailbox metadata needed by sharing state projection.
type MailboxEntry struct {
	// ID is the mailbox entry ID.
	ID int64
	// InviteID is the invitation the entry delivers.
	InviteID string
	// PeerID is the peer that redeemed the invitation.
	PeerID string
	// Status is the delivery status, such as pending.
	Status string
	// CreatedAt is the entry creation time.
	CreatedAt int64
	// AccountID is the account of the redeeming peer.
	AccountID string
	// EntityID is the entity label of the redeeming account.
	EntityID string
}

// equalMailboxEntry reports whether two mailbox entries hold the same values.
func equalMailboxEntry(a *MailboxEntry, b *MailboxEntry) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
