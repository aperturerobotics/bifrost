package nats

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	stream_netconn "github.com/aperturerobotics/bifrost/stream/netconn"
	"github.com/golang/protobuf/proto"
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
		_ = s.m.natsServer.HandleClientConnection(nconn)
	case NatsConnType_NatsConnType_ROUTER:
		s.m.natsServer.HandleRouterConnection(nconn, peerInfo.Pretty())
	default:
		// err := errors.Errorf("unknown nats conn type: %v", s.strmType.String())
		le.Warn("rejecting session with unknown nats conn type")
		return
	}

	<-ctx.Done()
	le.Debug("session closed")
}

// computeMessageID computes a message id for a packet
func computeMessageID(pkt *PubMessage) string {
	inner := bytes.Join([][]byte{
		pkt.GetSignature().GetSigData(),
		[]byte(pkt.GetFromPeerId()),
	}, nil)
	s := sha1.Sum(inner)
	return hex.EncodeToString(s[:])
}

// handlePublish handles incoming published packets
func (s *streamHandler) handlePublish(pkts []*PubMessage) {
	for _, pkt := range pkts {
		pid, err := peer.IDB58Decode(pkt.GetFromPeerId())
		if err != nil {
			s.le.WithError(err).Warnf("cannot decode peer id: %s", pkt.GetFromPeerId())
			continue
		}
		pubKey, err := pid.ExtractPublicKey()
		if err != nil {
			s.le.WithError(err).Warnf(
				"unable to extract pub key from peer id: %s",
				pkt.GetFromPeerId(),
			)
			continue
		}
		// validate signature
		sigOk, sigErr := pkt.GetSignature().VerifyWithPublic(pubKey, pkt.GetData())
		if sigErr != nil {
			s.le.WithError(sigErr).Warn("unable to validate signature, dropping")
			continue
		}
		if !sigOk {
			s.le.Warn("invalid signature, dropping")
			continue
		}
		pktInner := &PubMessageInner{}
		if err := proto.Unmarshal(pkt.GetData(), pktInner); err != nil {
			s.le.WithError(err).Warn("cannot unmarshal inner data")
			continue
		}
		chid := pktInner.GetChannel()
		if chid == "" {
			s.le.Warn("channel id in inner was empty")
			continue
		}
		/* TODO
		s.m.mtx.Lock()
		_, chOk := s.m.channels[chid]
		s.m.mtx.Unlock()
		if !chOk {
			s.le.Warnf("received message for non-subscribed channel %s", chid)
			continue
		}
		s.m.handleValidMessage(s.ctx, s.peerID, pkt, pktInner)
		*/
	}
}
