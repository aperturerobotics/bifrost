package pconn

import (
	"github.com/aperturerobotics/bifrost/stream"
)

// smuxAcceptPump accepts streams from the stream muxer.
// rawStreamsMtx should be locked by the caller
func (l *Link) smuxAcceptPump(initiator bool) {
	// Accept or create the initial coordination stream.
	var coordStrm stream.Stream
	var err error
	if !initiator {
		coordStrm, err = l.mux.OpenStream()
	} else {
		coordStrm, err = l.mux.AcceptStream()
	}
	l.coordStream = coordStrm
	l.rawStreamsMtx.Unlock()
	if err != nil {
		l.le.WithError(err).Warn("error opening coordination stream")
		_ = l.Close()
		return
	}

	// Start coordination processor
	go l.coordinationStreamPump(coordStrm)

	for {
		s, err := l.mux.AcceptStream()
		if err != nil {
			_ = l.Close()
			return
		}

		select {
		case <-l.ctx.Done():
			s.Close()
			return
		case l.acceptStreamCh <- &acceptedStream{
			stream: s,
			streamOpts: stream.OpenOpts{
				Reliable:  true,
				Encrypted: true,
			},
		}:
		}
	}
}
