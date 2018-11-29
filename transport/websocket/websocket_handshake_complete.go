package websocket

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/golang/protobuf/proto"
)

// handleCompleteHandshake handles a completed handshake.
func (u *Transport) handleCompleteHandshake(
	ctx context.Context,
	url string,
	conn net.Conn,
	result *identity.Result,
	initiator bool,
) {
	pid, _ := peer.IDFromPublicKey(result.Peer)
	le := u.le.
		WithField("remote-id", pid.Pretty()).
		WithField("remote-url", url)
	le.Info("handshake complete")

	edDat := result.ExtraData
	ed := &HandshakeExtraData{}
	if err := proto.Unmarshal(edDat, ed); err != nil {
		le.WithError(err).Warn("cannot unmarshal remote extra data")
	}

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
		u.peerID,
		ed.GetLocalTransportUuid(),
		result,
		result.Secret,
		conn,
		initiator,
		func() {
			go u.handleLinkLost(url, nlnk)
		},
	)
	u.links[url] = nlnk
	if u.handler != nil {
		u.handler.HandleLinkEstablished(nlnk)
	}
}
