package websocket

import (
	"net"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
)

// handleCompleteHandshake handles a completed handshake.
func (u *Transport) handleCompleteHandshake(
	url string,
	conn net.Conn,
	result *identity.Result,
	initiator bool,
) {
	ctx := u.ctx
	pid, _ := peer.IDFromPublicKey(result.Peer)
	le := u.le.
		WithField("remote-id", pid.Pretty()).
		WithField("remote-url", url)
	le.Info("handshake complete")

	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	// TODO; re-configure link for new secret rather than closing it.
	// TODO: find any peers with this ID and userp
	if l, ok := u.links[url]; ok {
		le.
			Debug("userping old session with peer")
		l.Close()
	}

	var nlnk *Link
	nlnk = NewLink(
		ctx,
		le,
		url,
		u.GetUUID(),
		result,
		result.Secret,
		conn,
		initiator,
		func() {
			le.Debug("handleLinkLost()")
			go u.handleLinkLost(url, nlnk)
		},
	)
	u.links[url] = nlnk
	if u.handler != nil {
		u.handler.HandleLinkEstablished(nlnk)
	}
}
