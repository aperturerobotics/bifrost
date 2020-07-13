package nats

// TODO - WIP

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/pubsub"

	nats_server "github.com/nats-io/nats-server/v2/server"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// maxMessageSize constrains the message buffer allocation size.
// currently set to 2MB
const maxMessageSize = 2000000

const (
	NatsRouterID = protocol.ID("nats.io/2/router") // nats 2.0 router
	NatsClientID = protocol.ID("nats.io/2/client") // nats 2.0 client API
)

// ProtocolIDToStreamType converts a protocol ID to a conn type.
func ProtocolIDToStreamType(id protocol.ID) NatsConnType {
	switch id {
	case NatsRouterID:
		return NatsConnType_NatsConnType_ROUTER
	case NatsClientID:
		return NatsConnType_NatsConnType_CLIENT
	default:
		return NatsConnType_NatsConnType_UNKNOWN
	}
}

// TODO: move this code into a common "next-hop stream solicitation" package

// Nats implements the nats router.
type Nats struct {
	// conf is the config
	conf *Config
	// le is the logger
	le *logrus.Entry
	// peer contains the peer we are using
	peer peer.Peer
	// handler is the pubsub handler
	handler pubsub.PubSubHandler
	// wakeCh wakes the execute loop
	wakeCh chan struct{}
	// publishCh is for publishing messages
	publishCh chan *publishChMsg
	// natsServer is the embedded nats server.
	natsServer *nats_server.Server

	mtx         sync.Mutex
	incSessions []*streamHandler
}

// publishChMsg is a message queued for publishing
type publishChMsg struct {
	msg         *PubMessage
	prevHopPeer peer.ID
	channelID   string
}

// NewNats constructs a new Nats PubSub router.
func NewNats(
	ctx context.Context,
	le *logrus.Entry,
	handler pubsub.PubSubHandler,
	cc *Config,
	peer peer.Peer,
) (pubsub.PubSub, error) {
	kpair, err := NewKeyPair(peer.GetPrivKey(), peer.GetPubKey())
	if err != nil {
		return nil, err
	}
	peerID := peer.GetPeerID()
	serverName := peerID.Pretty()

	if peer.GetPrivKey() == nil {
		return nil, errors.New("peer must have priv key")
	}

	clusterName := cc.GetClusterName()
	if clusterName == "" {
		clusterName = string(NatsRouterID)
	}

	serverOpts := &nats_server.Options{
		ServerName: serverName,
		Cluster: nats_server.ClusterOpts{
			Name: clusterName,
		},
		Debug: true,
		Trace: true,
		Logger: le.WithFields(logrus.Fields{
			"peer-id": serverName,
		}),
	}
	if err := cc.ApplyOptions(serverOpts); err != nil {
		return nil, err
	}

	// Create nats server with Aperture fork
	natsServer, err := nats_server.NewServer(
		serverOpts,
		kpair,
	)
	if err != nil {
		return nil, err
	}

	return &Nats{
		le:         le,
		conf:       cc,
		handler:    handler,
		peer:       peer,
		natsServer: natsServer,

		wakeCh:    make(chan struct{}, 1),
		publishCh: make(chan *publishChMsg, 16),
	}, nil
}

// Execute executes the PubSub routines.
func (m *Nats) Execute(ctx context.Context) error {
	m.le.Debug("nats router starting")

	go m.natsServer.Start()
	defer m.natsServer.WaitForShutdown()
	defer m.natsServer.Shutdown()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-m.wakeCh:
		}

		m.mtx.Lock()
		incSess := m.incSessions
		m.incSessions = nil

		for _, sess := range incSess {
			sess.ctx, sess.ctxCancel = context.WithCancel(ctx)
			go sess.executeSession()
		}
		m.mtx.Unlock()
	}
}

// execPublish executes publishing a message
func (m *Nats) execPublish(prevHopPeerID peer.ID, pubMsg *publishChMsg) {
	// chid := pubMsg.channelID
	// tosend := make(map[pubsub.PeerLinkTuple]struct{})
	/*
		m.mtx.Lock()
		if peerChannels, ok := m.peerChannels[chid]; ok {
			for p := range peerChannels {
				tosend[p] = struct{}{}
			}
		}
		for pid := range tosend {
			pidPretty := pid.PeerID.Pretty()
			if pidPretty == pubMsg.msg.GetFromPeerId() ||
				pid.PeerID == prevHopPeerID {
				continue
			}

			peer, ok := m.peers[pid]
			if ok {
				peer.writePacket(pkt)
			}
		}
		m.mtx.Unlock()
	*/
}

// AddPeerStream adds a negotiated peer stream.
// Two streams will be negotiated, one outgoing, one incoming.
// The pubsub should communicate over the stream.
func (m *Nats) AddPeerStream(
	tpl pubsub.PeerLinkTuple,
	initiator bool,
	mstrm link.MountedStream,
) {
	le := m.le.WithField("peer", tpl.PeerID.Pretty())
	if !mstrm.GetOpenOpts().Encrypted || !mstrm.GetOpenOpts().Reliable {
		le.Warn("rejecting unencrypted or unreliable pubsub stream")
		mstrm.GetStream().Close()
		return
	}
	protocolID := mstrm.GetProtocolID()
	streamType := ProtocolIDToStreamType(protocolID)
	if streamType == NatsConnType_NatsConnType_UNKNOWN {
		le.
			WithField("protocol-id", protocolID).
			WithField("stream-type", streamType.String()).
			Warn("rejecting unknown protocol id")
		mstrm.GetStream().Close()
		return
	}

	sh := &streamHandler{
		m:         m,
		le:        le,
		tpl:       tpl,
		peerID:    mstrm.GetPeerID(),
		mstrm:     mstrm,
		initiator: initiator,
		strmType:  streamType,
	}
	m.mtx.Lock()
	m.incSessions = append(m.incSessions, sh)
	m.mtx.Unlock()
	m.wake()
}

// Close closes the pubsub.
func (m *Nats) Close() {
	m.natsServer.Shutdown()
	m.natsServer.WaitForShutdown()
	m.mtx.Lock()
	for _, s := range m.incSessions {
		if s.mstrm != nil {
			s.mstrm.GetStream().Close()
		}
	}
	m.incSessions = nil
	m.mtx.Unlock()
}

// handleValidMessage handles a valid message, repeating it to other peers and handlers.
func (m *Nats) handleValidMessage(
	ctx context.Context,
	prevHopPeer peer.ID,
	pkt *PubMessage,
	pktInner *PubMessageInner,
) {
	channelID := pktInner.GetChannel()
	msgId := computeMessageID(pkt)

	pid, err := peer.IDB58Decode(pkt.GetFromPeerId())
	if err != nil {
		return
	}
	msg := newMessage(pid, pktInner)
	_ = channelID
	_ = msg
	_ = msgId
	// TODO
	/*
		m.mtx.Lock()
		subs := m.channels[channelID]
		for sub := range subs {
			ss := sub
			go func() {
				ss.mtx.Lock()
				for s := range ss.handlers {
					s.cb(msg)
				}
				ss.mtx.Unlock()
			}()
		}
		m.mtx.Unlock()
		select {
		case m.publishCh <- &publishChMsg{
			msg:         pkt,
			channelID:   channelID,
			prevHopPeer: prevHopPeer,
		}:
		case <-ctx.Done():
		}
	*/
}

// wake wakes the controller
func (m *Nats) wake() {
	select {
	case m.wakeCh <- struct{}{}:
	default:
	}
}

// AddSubscription adds a channel subscription, returning a subscription handle.
//
// Uses the router peer private key.
//
// An alternate approach is to use a client connection.
func (n *Nats) AddSubscription(ctx context.Context, channelID string) (pubsub.Subscription, error) {
	return nil, errors.New("TODO implement nats add subscription")
}

// _ is a type assertion
var _ pubsub.PubSub = ((*Nats)(nil))
