package nats

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	stream_netconn "github.com/aperturerobotics/bifrost/stream/netconn"
	"github.com/sirupsen/logrus"
)

// streamHandler is a remote floodsub peer with a stream.
type streamHandler struct {
	tpl       pubsub.PeerLinkTuple
	m         *Nats
	initiator bool
	le        *logrus.Entry
	peerID    peer.ID
	mstrm     link.MountedStream
	strmType  NatsConnType

	ctx       context.Context
	ctxCancel context.CancelFunc
}

// executeSession executes the stream session.
func (s *streamHandler) executeSession() {
	ctx := s.ctx
	defer s.ctxCancel()
	defer s.mstrm.GetStream().Close()

	// c contains the nats client
	le := s.le.
		WithField("initiator", s.initiator).
		WithField("nats-stream-type", s.strmType.String())
	le.Debug("executing session")
	nconn := stream_netconn.NewNetConn(s.mstrm)
	peerInfo := s.mstrm.GetPeerID()
	switch s.strmType {
	case NatsConnType_NatsConnType_CLIENT:
		_ = s.m.natsServer.HandleClientConnection(nconn, peerInfo.String())
	case NatsConnType_NatsConnType_ROUTER:
		s.m.natsServer.HandleRouterConnection(nconn, peerInfo.String())
	default:
		// err := errors.Errorf("unknown nats conn type: %v", s.strmType.String())
		le.Warn("rejecting session with unknown nats conn type")
		return
	}

	<-ctx.Done()
	le.Debug("session closed")
}
