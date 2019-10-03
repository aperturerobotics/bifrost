package floodsub

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	stream_packet "github.com/aperturerobotics/bifrost/stream/packet"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// streamHandler is a remote floodsub peer with a stream.
type streamHandler struct {
	tpl       pubsub.PeerLinkTuple
	m         *FloodSub
	initiator bool
	stream    *stream_packet.Session
	le        *logrus.Entry
	packetCh  chan *Packet
	peerID    peer.ID

	ctx       context.Context
	ctxCancel context.CancelFunc
}

// writePacket writes a packet.
func (s *streamHandler) writePacket(pkt *Packet) {
	select {
	case s.packetCh <- pkt:
	case <-s.ctx.Done():
	}
}

// executeSession executes the stream session.
func (s *streamHandler) executeSession() error {
	ctx := s.ctx
	defer s.stream.Close()
	defer s.ctxCancel()
	go s.readPump(ctx)

	// le := s.le.WithField("initiator", s.initiator)
	// le.Info("executing session")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case pkt := <-s.packetCh:
			if err := s.stream.SendMsg(pkt); err != nil {
				return err
			}
		}
	}
}

// processPacket processes the incoming packet.
func (s *streamHandler) processPacket(msg *Packet) {
	if subs := msg.GetSubscriptions(); len(subs) != 0 {
		s.handleSubscriptions(subs)
	}
	if pubs := msg.GetPublish(); len(pubs) != 0 {
		s.handlePublish(pubs)
	}
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
		s.m.mtx.Lock()
		_, chOk := s.m.channels[chid]
		s.m.mtx.Unlock()
		if !chOk {
			s.le.Warnf("received message for non-subscribed channel %s", chid)
			continue
		}
		s.m.handleValidMessage(s.ctx, s.peerID, pkt, pktInner)
	}
}

// handleSubscriptions processes subscription packet data
func (s *streamHandler) handleSubscriptions(subs []*SubscriptionOpts) {
	s.m.mtx.Lock()
	defer s.m.mtx.Unlock()

	for _, sub := range subs {
		chid := sub.GetChannelId()
		if chid == "" {
			continue
		}
		le := s.le.WithField("channel-id", chid)
		if sub.GetSubscribe() {
			cm, ok := s.m.peerChannels[chid]
			if !ok {
				cm = make(map[pubsub.PeerLinkTuple]struct{})
				s.m.peerChannels[chid] = cm
			}
			if _, ok := cm[s.tpl]; !ok {
				le.Debug("peer subscribed to channel")
				cm[s.tpl] = struct{}{}
			}
		} else {
			tm, ok := s.m.peerChannels[chid]
			if !ok {
				continue
			}
			if _, ok := tm[s.tpl]; ok {
				le.Debug("peer unsubscribed from channel")
				delete(tm, s.tpl)
			}
			if len(tm) == 0 {
				delete(s.m.peerChannels, chid)
			}
		}
	}
}

// readPump reads messages from the stream.
func (s *streamHandler) readPump(ctx context.Context) {
	defer s.ctxCancel()

	msg := &Packet{}
	for {
		if err := s.stream.RecvMsg(msg); err != nil {
			if err != io.EOF && err != context.Canceled && err.Error() != "broken pipe" && err.Error() == "NO_ERROR" {
				s.le.WithError(err).Warn("error receiving message")
			} else {
				s.le.Debug("session reader exiting")
			}
			return
		}
		s.le.
			WithField("subscription-count", len(msg.GetSubscriptions())).
			WithField("publish-count", len(msg.GetPublish())).
			Debug("received message from peer")
		// process message
		s.processPacket(msg)
		msg.Reset()
	}
}
