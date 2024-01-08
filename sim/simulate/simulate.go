package simulate

import (
	"context"
	"sync"

	link_establish_controller "github.com/aperturerobotics/bifrost/link/establish"
	bpeer "github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/sirupsen/logrus"
)

// Simulator manages state for all simulated machines.
type Simulator struct {
	// ctx is the context
	ctx context.Context
	// ctxCancel cancels the context
	ctxCancel context.CancelFunc
	// le is the logger
	le *logrus.Entry
	// mtx guards below fields
	mtx sync.Mutex
	// peers contains running peer nodes
	peers map[string]*Peer
}

// NewSimulator constructs a new sim.
func NewSimulator(
	ctx context.Context,
	le *logrus.Entry,
	grp *graph.Graph,
) (*Simulator, error) {
	s := &Simulator{
		le:    le,
		peers: make(map[string]*Peer),
	}
	s.ctx, s.ctxCancel = context.WithCancel(ctx)

	// Instantiate the nodes
	allNodes := grp.AllNodes()
	le.Debugf("processing %d nodes in graph", len(allNodes))
	for _, node := range allNodes {
		peer, isPeer := node.(*graph.Peer)
		if isPeer {
			peerID := peer.GetPeerID()
			peerIDStr := peerID.String()
			if _, epOk := s.peers[peerIDStr]; epOk {
				continue
			}

			pushedPeer, err := s.pushPeer(peer)
			if err != nil {
				s.ctxCancel()
				return nil, err
			}

			// get the list of linked peers
			linkedPeers := peer.GetLinkedPeers(grp)
			var linkedPeerIDs []bpeer.ID
			le.Debugf("peer %s has %d linked peers", peerIDStr, len(linkedPeers))

			// add each linked peer
			for _, lpeer := range linkedPeers {
				lpeerPeerIDStr := lpeer.GetPeerID().String()
				linkedPeerIDs = append(linkedPeerIDs, lpeer.GetPeerID())
				op, ok := s.peers[lpeerPeerIDStr]
				if !ok {
					continue
				}
				op.inproc.ConnectToInproc(ctx, pushedPeer.inproc)
				pushedPeer.inproc.ConnectToInproc(s.ctx, op.inproc)
				le.Debugf("adding in-memory link from %s from %s", lpeerPeerIDStr, peerIDStr)
				pushedPeer.transportController.PushStaticPeer(lpeerPeerIDStr, &dialer.DialerOpts{
					Address: op.inproc.LocalAddr().String(),
				})
				op.transportController.PushStaticPeer(peerIDStr, &dialer.DialerOpts{
					Address: pushedPeer.inproc.LocalAddr().String(),
				})
			}

			// push the peer establish controller to trigger the link
			_, err = pushedPeer.testbed.Bus.AddController(
				s.ctx,
				link_establish_controller.NewController(
					pushedPeer.testbed.Bus,
					pushedPeer.le,
					linkedPeerIDs,
					"",
				),
				nil,
			)
			if err != nil {
				s.ctxCancel()
				return nil, err
			}

			continue
		}
	}

	return s, nil
}

// pushPeer creates and starts a peer in the simulator.
// expects caller to hold lock on mtx
func (s *Simulator) pushPeer(peer *graph.Peer) (*Peer, error) {
	p, err := newPeer(s.ctx, s.le, peer)
	if err != nil {
		return nil, err
	}
	peerIDStr := p.GetPeerID().String()
	s.peers[peerIDStr] = p
	return p, nil
}

// GetPeerByID returns a peer by ID.
func (s *Simulator) GetPeerByID(id bpeer.ID) *Peer {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.peers[id.String()]
}

// Close closes the simulator.
func (s *Simulator) Close() {
	s.ctxCancel()
}
