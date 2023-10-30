package dialer

import (
	"context"
	"errors"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	bo "github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

// Dialer manages a transport dialer.
type Dialer struct {
	// le is the logger
	le *logrus.Entry
	// tptDialer is the transport dialer.
	tptDialer transport.TransportDialer
	// backoff is the dialer backoff
	backoff bo.BackOff
	// peerID is the peer id
	peerID peer.ID
	// address is the dial address
	address string
}

// NewDialer constructs a new Dialer
func NewDialer(
	le *logrus.Entry,
	tptDialer transport.TransportDialer,
	opts *DialerOpts,
	peerID peer.ID,
	address string,
) *Dialer {
	return &Dialer{
		le: le.WithField("dial-peer-id", peerID.String()).
			WithField("dial-peer-addr", address),
		tptDialer: tptDialer,
		backoff:   opts.GetBackoff().Construct(),
		peerID:    peerID,
		address:   address,
	}
}

// Execute executes the dialer, with backoff.
func (d *Dialer) Execute(ctx context.Context) error {
	for {
		d.le.Debug("attempting to dial peer")
		fatal, err := d.tptDialer.DialPeer(ctx, d.peerID, d.address)
		if err == nil {
			d.backoff.Reset()
			return nil
		}

		bo := d.backoff.NextBackOff()
		if err != nil {
			if err == context.Canceled {
				return err
			}

			if fatal {
				d.le.WithError(err).Warn("dialer errored fatally")
				return err
			}

			d.le.
				WithError(err).
				WithField("backoff", bo.String()).
				Warn("dialer errored")
		}

		if bo == -1 {
			return errors.New("dial backoff max duration exceeded")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(bo):
		}
	}
}
