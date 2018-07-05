//+build !js

package websocket

import (
	"context"
	"io"

	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/gorilla/websocket"
)

// inflightHandshake is an on-going handshake.
type inflightHandshake struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	url       string
	conn      *websocket.Conn
}

// pushHandshaker builds a new handshaker for the address.
// it is expected that handshakesMtx is locked before calling pushHandshaker
// if conn is nil, will dial
func (u *Transport) pushHandshaker(
	ctx context.Context,
	url string,
	conn *websocket.Conn,
) (*inflightHandshake, error) {
	nctx, nctxCancel := context.WithTimeout(ctx, handshakeTimeout)
	hs := &inflightHandshake{ctxCancel: nctxCancel, url: url, conn: conn, ctx: nctx}
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

	if hs.conn == nil {
		ule.Debug("dialing")
		conn, _, err := websocket.DefaultDialer.Dial(hs.url, nil)
		if err != nil {
			ule.WithError(err).Warn("websocket dial errored")
			return
		}
		hs.conn = conn
	}

	ule.Debug("handshaking")
	hns, err := s2s.NewHandshaker(
		u.privKey,
		nil,
		func(data []byte) error {
			return hs.conn.WriteMessage(websocket.BinaryMessage, data)
		}, nil, nil,
	)
	if err != nil {
		ule.WithError(err).Warn("error building handshaker")
		hs.conn.Close()
		return
	}

	go func() {
		for {
			mt, buf, err := hs.conn.ReadMessage()
			if err != nil {
				if err != io.EOF {
					ule.
						WithError(err).
						Warn("error reading while handshaking")
				}
				return
			}

			if mt != websocket.BinaryMessage {
				continue
			}

			if !hns.Handle(buf) {
				return
			}
		}
	}()

	initiator := false
	res, err := hns.Execute(ctx, initiator)
	if err != nil {
		if err == context.Canceled {
			return
		}

		ule.WithError(err).Warn("error handshaking")
		_ = hs.conn.Close()
		return
	}

	select {
	case <-ctx.Done():
		ule.WithError(ctx.Err()).Warn("handshake succeeded but ctx canceled")
		hs.conn.Close()
		return
	default:
	}

	ule.Debug("handshake complete")
	u.handleCompleteHandshake(hs.url, newWsConn(hs.conn), res, initiator)
}
