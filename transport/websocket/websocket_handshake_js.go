//+build js

package websocket

import (
	"context"
	"io"

	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/gopherjs/websocket"
)

// inflightHandshake is an on-going handshake.
type inflightHandshake struct {
	ctxCancel context.CancelFunc
	url       string
}

// pushHandshaker builds a new handshaker for the address.
// it is expected that handshakesMtx is locked before calling pushHandshaker
// if conn is set will assume is a listener
func (u *Transport) pushHandshaker(
	ctx context.Context,
	url string,
) (*inflightHandshake, error) {
	nctx, nctxCancel := context.WithTimeout(ctx, handshakeTimeout)
	hs := &inflightHandshake{ctxCancel: nctxCancel, url: url}
	if old, ok := u.handshakes[url]; ok && old.ctxCancel != nil {
		old.ctxCancel()
	}
	u.handshakes[url] = hs

	go u.processHandshake(nctx, hs)
	return hs, nil
}

// processHandshake processes an in-flight handshake.
func (u *Transport) processHandshake(ctx context.Context, hs *inflightHandshake) {
	ule := u.le.WithField("url", hs.url)

	defer func() {
		u.handshakesMtx.Lock()
		ohs := u.handshakes[hs.url]
		if ohs == hs {
			delete(u.handshakes, hs.url)
		}
		u.handshakesMtx.Unlock()
	}()

	ule.Debug("dialing")
	var err error
	conn, err := websocket.Dial(hs.url)
	if err != nil {
		ule.WithError(err).Warn("websocket dial errored")
		return
	}

	ule.Debug("handshaking")
	hns, err := s2s.NewHandshaker(
		u.privKey,
		nil,
		func(data []byte) error {
			_, err := conn.Write(data)
			return err
		}, nil, nil,
	)
	if err != nil {
		ule.WithError(err).Warn("error building handshaker")
		conn.Close()
		return
	}

	go func() {
		var buf [1300]byte
		for {
			n, err := conn.Read(buf[:])
			if err != nil {
				if err != io.EOF {
					ule.
						WithError(err).
						Warn("error reading while handshaking")
				}
				return
			}

			if !hns.Handle(buf[:n]) {
				return
			}
		}
	}()

	initiator := true
	res, err := hns.Execute(ctx, initiator)
	if err != nil {
		if err == context.Canceled {
			return
		}

		ule.WithError(err).Warn("error handshaking")
		conn.Close()
		return
	}

	select {
	case <-ctx.Done():
		ule.WithError(ctx.Err()).Warn("handshake succeeded but ctx canceled")
		conn.Close()
		return
	default:
	}

	ule.Debug("handshake complete")
	u.handleCompleteHandshake(hs.url, conn, res, initiator)
}
