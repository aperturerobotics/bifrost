package provider_spacewave

import (
	"context"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/pairing"
	"github.com/s4wave/spacewave/core/transport"
)

// GetPairingEngine returns this unlocked Session's account enrollment owner.
func (s *Session) GetPairingEngine() (*pairing.Engine, error) {
	s.pairingMu.Lock()
	defer s.pairingMu.Unlock()
	if s.sessionPriv == nil || s.lifecycleCtx.Err() != nil {
		return nil, errors.New("Session is locked or closed")
	}
	if s.pairingEngine != nil {
		return s.pairingEngine, nil
	}
	ctx, cancel := context.WithCancel(s.lifecycleCtx)
	a := s.tkr.a
	engine, err := pairing.NewEngine(ctx, a.le, a.p.b, s.sessionPriv, a, s.GetPairingTransport)
	if err != nil {
		cancel()
		return nil, err
	}
	s.pairingEngine, s.pairingCancel = engine, cancel
	return engine, nil
}

func (s *Session) clearPairingEngine() {
	s.pairingMu.Lock()
	defer s.pairingMu.Unlock()
	if s.pairingCancel != nil {
		s.pairingCancel()
	}
	s.pairingEngine, s.pairingCancel = nil, nil
}

// GetPairingTransport uses this Session's ordinary transport composition.
// Starting pairing explicitly enables the connection policy it requires.
func (s *Session) GetPairingTransport(ctx context.Context, relay pairing.Relay) (*transport.SessionTransport, error) {
	if !s.GetDirectP2PEnabled() {
		if err := s.SetDirectP2PEnabled(ctx, true); err != nil {
			return nil, err
		}
	}
	a := s.tkr.a
	for {
		var st *transport.SessionTransport
		var wait <-chan struct{}
		a.transportBcast.HoldLock(func(_ func(), getWait func() <-chan struct{}) {
			wait = getWait()
			if state := a.sessionTransports[s.tkr.id]; state != nil {
				st = state.transport
			}
		})
		if st != nil {
			return st, st.AwaitReady(ctx)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-wait:
		}
	}
}
