package provider_spacewave

import (
	"context"
	"slices"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/pairing"
	"github.com/s4wave/spacewave/core/provider"
	provider_migration "github.com/s4wave/spacewave/core/provider/migration"
)

var errAccountTransitionPending = errors.New("this Session is moving to its destination account")

// observeAccountTransition separates the server's new attachment from this
// account's cached data. A later transition may supersede an offline attachment.
func (a *ProviderAccount) observeAccountTransition(transition *provider.AccountTransition, accountID, peerID string) bool {
	if transition == nil || transition.Validate() != nil || !slices.Contains(transition.GetSessionPeerIds(), peerID) {
		return false
	}
	if transition.GetSourceEndpoint() != a.p.endpoint {
		return false
	}
	if accountID == a.accountID && transition.GetSource().GetProviderAccountId() != a.accountID {
		return false
	}
	if accountID != a.accountID && (transition.GetDestinationEndpoint() != a.p.endpoint || transition.GetDestination().GetProviderAccountId() != accountID) {
		return false
	}
	a.accountBcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if !a.state.transition.EqualVT(transition) {
			a.state.transition = transition.CloneVT()
			broadcast()
		}
	})
	return true
}

// watchAccountTransition follows cloud authorization under the unlocked Session
// lifetime. The initial read also covers a Session that was offline at commit.
func (s *Session) watchAccountTransition(ctx context.Context) error {
	a := s.tkr.a
	client, err := a.migrationClient(s.GetPrivKey())
	if err != nil {
		return err
	}
	info, err := client.GetAccountInfo(ctx)
	if err != nil {
		return err
	}
	a.observeAccountTransition(info.GetTransition(), info.GetAccountId(), s.GetPeerId().String())
	for {
		var transition *provider.AccountTransition
		var changed <-chan struct{}
		a.accountBcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
			transition = a.state.transition
			changed = getWait()
		})
		if transition != nil {
			s.pairingMu.Lock()
			engine := s.pairingEngine
			s.pairingMu.Unlock()
			if engine != nil {
				for {
					snapshot, changed := engine.Snapshot()
					if snapshot.Status == pairing.StatusBothConfirmed && snapshot.Choice.Merging() {
						// The pairing owner already committed this attachment.
						<-ctx.Done()
						return ctx.Err()
					}
					if snapshot.Status != pairing.StatusEnrolling || !snapshot.Choice.Merging() {
						break
					}
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-changed:
					}
				}
			}
			p, releaseProvider, err := provider.ExLookupProvider(ctx, a.p.b, transition.GetDestination().GetProviderId(), false, nil)
			if err != nil {
				return err
			}
			defer releaseProvider.Release()
			if p == nil {
				return errors.New("destination provider is not configured")
			}
			account, releaseAccount, err := p.AccessProviderAccount(ctx, transition.GetDestination().GetProviderAccountId(), nil)
			if err != nil {
				return err
			}
			defer releaseAccount()
			target, ok := account.(provider_migration.Account)
			if !ok {
				return errors.New("destination provider cannot accept this returning Session")
			}
			next, err := target.AttachMigratedSession(ctx, s, transition)
			if err != nil {
				return err
			}
			if err := provider_migration.RebindSession(ctx, s, next); err != nil {
				return err
			}
			<-ctx.Done()
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}
