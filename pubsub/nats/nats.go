package nats

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"

	nats_server "github.com/nats-io/nats-server/v2/server"
	nats_client "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
	// natsServer is the embedded nats server.
	natsServer *nats_server.Server

	mtx         sync.Mutex
	incSessions []*streamHandler
	// natsClients contains all active nats clients keyed by peer id
	natsClients map[string]*natsClient
}

// NewNats constructs a new Nats PubSub router.
func NewNats(
	ctx context.Context,
	le *logrus.Entry,
	handler pubsub.PubSubHandler,
	cc *Config,
	peer peer.Peer,
) (pubsub.PubSub, error) {
	if peer == nil {
		return nil, errors.New("nats server requires a peer with a private key")
	}
	peerPrivKey, err := peer.GetPrivKey(ctx)
	if err != nil {
		return nil, err
	}
	kpair, err := NewKeyPair(peerPrivKey, peer.GetPubKey())
	if err != nil {
		return nil, err
	}
	peerID := peer.GetPeerID()
	serverName := peerID.String()

	clusterName := cc.GetClusterName()
	if clusterName == "" {
		clusterName = string(NatsRouterID)
	}

	serverOpts := &nats_server.Options{
		Cluster:    nats_server.ClusterOpts{Name: clusterName},
		Logger:     le.WithField("peer-id", serverName),
		ServerName: serverName,

		CustomClientAuthentication: newClientAuth(),
		CustomRouterAuthentication: newRouterAuth(),
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

		wakeCh:      make(chan struct{}, 1),
		natsClients: make(map[string]*natsClient),
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

// AddPeerStream adds a negotiated peer stream.
// Two streams will be negotiated, one outgoing, one incoming.
// The pubsub should communicate over the stream.
func (m *Nats) AddPeerStream(tpl pubsub.PeerLinkTuple, initiator bool, mstrm link.MountedStream) {
	le := m.le.WithField("peer", tpl.PeerID.String())
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

// BuildClient builds a client for the nats server, creating a client connection.
//
// Note: the servers list & dialer will be overwritten.
func (n *Nats) BuildClient(ctx context.Context, privKey crypto.PrivKey, opts ...nats_client.Option) (*nats_client.Conn, error) {
	nk, err := NewKeyPair(privKey, privKey.GetPublic())
	if err != nil {
		return nil, err
	}
	nkPub, err := nk.PublicKey()
	if err != nil {
		return nil, err
	}
	copts := nats_client.GetDefaultOptions()
	for _, opt := range opts {
		if err := opt(&copts); err != nil {
			return nil, err
		}
	}
	copts.CustomDialer = newLocalNatsDialer(n, nk)
	copts.Servers = []string{localNatsAddress}
	if err := nats_client.Nkey(nkPub, nats_client.SignatureHandler(nk.Sign))(&copts); err != nil {
		return nil, err
	}
	return copts.Connect()
}

// AddSubscription adds a channel subscription, returning a subscription handle.
//
// Uses the router peer private key.
//
// An alternate approach is to use a client connection.
func (n *Nats) AddSubscription(ctx context.Context, privKey crypto.PrivKey, channelID string) (pubsub.Subscription, error) {
	n.mtx.Lock()
	nc, ncRel, err := n.getOrBuildClient(ctx, privKey)
	n.mtx.Unlock()
	if err != nil {
		return nil, err
	}

	nsub, err := nc.Conn.SubscribeSync(channelID)
	if err != nil {
		ncRel()
		return nil, err
	}

	return newSubscription(ctx, n, nc, ncRel, nsub, privKey, channelID), nil
}

// GetOrBuildCommonClient returns the common nats client.
func (n *Nats) GetOrBuildCommonClient(ctx context.Context) (*nats_client.Conn, error) {
	npeerID := n.peer.GetPeerID()
	npeerPriv, err := n.peer.GetPrivKey(ctx)
	if err != nil {
		return nil, err
	}
	npeerString := npeerID.String()

	n.mtx.Lock()
	defer n.mtx.Unlock()

	// we create the common client with 1 ref that is never released.
	nc, ncOk := n.natsClients[npeerString]
	if ncOk {
		return nc.Conn, nil
	}
	nc, ncRel, err := n.getOrBuildClient(ctx, npeerPriv)
	if err != nil {
		return nil, err
	}
	_ = ncRel // never release the common client
	return nc.Conn, nil
}

// getOrBuildClient gets or builds a client adding a reference.
// caller must lock mtx
func (n *Nats) getOrBuildClient(ctx context.Context, privKey crypto.PrivKey) (*natsClient, func(), error) {
	npeer, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}
	npeerString := npeer.String()
	nc := n.natsClients[npeerString]
	if nc != nil {
		if !nc.Conn.IsClosed() {
			return nc, nc.addRef(), nil
		}

		nc = nil
		delete(n.natsClients, npeerString)
	}

	nconn, err := n.BuildClient(ctx, privKey)
	if err != nil {
		return nil, nil, err
	}
	nc = newNatsClient(npeer, nconn)
	n.natsClients[npeerString] = nc
	return nc, nc.addRef(), nil
}

// SubscribeSync passes through to the nats client SubscribeSync call.
func (n *Nats) SubscribeSync(
	ctx context.Context,
	channelID string,
) (*nats_client.Subscription, *nats_client.Conn, error) {
	// TODO: do we need to augment this?
	nclient, err := n.GetOrBuildCommonClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	sub, err := nclient.SubscribeSync(channelID)
	if err != nil {
		return nil, nclient, err
	}
	return sub, nclient, nil
}

// Close closes the pubsub.
func (m *Nats) Close() {
	m.mtx.Lock()
	for _, s := range m.incSessions {
		if s.mstrm != nil {
			s.mstrm.GetStream().Close()
		}
	}
	m.incSessions = nil
	for id, client := range m.natsClients {
		client.Close()
		delete(m.natsClients, id)
	}
	m.mtx.Unlock()
}

// wake wakes the controller
func (m *Nats) wake() {
	select {
	case m.wakeCh <- struct{}{}:
	default:
	}
}

// _ is a type assertion
var _ pubsub.PubSub = ((*Nats)(nil))
