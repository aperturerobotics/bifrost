package provider_local

import (
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/net/peer"
)

// ConfirmPairingError identifies an unavailable or mismatched enrollment result.
type ConfirmPairingError string

const (
	// ErrPairingExchangeMissing means no pairing exchange exists.
	ErrPairingExchangeMissing ConfirmPairingError = "pairing exchange is missing"
	// ErrPairingExchangeUnconfirmed means account enrollment has not completed.
	ErrPairingExchangeUnconfirmed ConfirmPairingError = "account enrollment is not complete"
	// ErrPairingExchangePeerMismatch means the result belongs to another peer.
	ErrPairingExchangePeerMismatch ConfirmPairingError = "pairing exchange peer does not match"
)

// Error returns the pairing result failure.
func (e ConfirmPairingError) Error() string {
	return string(e)
}

// GetPairingResult returns the durably enrolled Session on the receiving client.
// The offering client returns nil. Repeated reads are safe after an RPC retry;
// authorization and data transfer belong to the bilateral exchange itself.
func (a *ProviderAccount) GetPairingResult(remotePeer peer.ID) (*session.SessionRef, error) {
	var ref *session.SessionRef
	var err error
	a.pairingBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		switch {
		case a.pairing == nil:
			err = ErrPairingExchangeMissing
		case a.pairing.remotePeerID != remotePeer:
			err = ErrPairingExchangePeerMismatch
		case a.pairing.status != PairingStatusBothConfirmed:
			err = ErrPairingExchangeUnconfirmed
		case !a.pairing.offering && a.pairing.enrolledSession == nil:
			err = ErrPairingExchangeUnconfirmed
		default:
			ref = a.pairing.enrolledSession.CloneVT()
		}
	})
	return ref, err
}
