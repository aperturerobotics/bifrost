package provider_local

import (
	"context"
	"maps"
	"sync"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/core/transport"
	"github.com/s4wave/spacewave/net/peer"
)

// staticUDP is the one UDP transport a P2P sync generation runs to dial peers
// at static addresses: those the listed objects carry, and the owners of
// invitations being joined before their objects are listed.
type staticUDP struct {
	// mtx guards the fields below and serializes restarts.
	mtx sync.Mutex
	// listed holds the addresses the listed objects carry.
	listed map[peer.ID]string
	// joining holds the owner addresses joins added. A listed address for the
	// same peer takes precedence.
	joining map[peer.ID]string
	// peers is the set the running transport dials, nil when none runs.
	peers map[peer.ID]string
	// cancel stops the running transport.
	cancel context.CancelFunc
}

// listedUDPPeers collects the static UDP address of each peer the listed
// objects name. An unparsable peer ID is skipped.
func (a *ProviderAccount) listedUDPPeers(list *sobject.SharedObjectList) map[peer.ID]string {
	var peers map[peer.ID]string
	for _, entry := range list.GetSharedObjects() {
		for id, addr := range entry.GetUdpPeerAddrs() {
			remote, err := peer.IDB58Decode(id)
			if err != nil {
				a.le.WithError(err).WithField("peer-id", id).Warn("skipping listed udp peer")
				continue
			}
			if peers == nil {
				peers = make(map[peer.ID]string)
			}
			peers[remote] = addr
		}
	}
	return peers
}

// reconcileStaticUDP dials the addresses the listed objects carry. A transport
// that fails to start is retried on the next list change.
func (a *ProviderAccount) reconcileStaticUDP(state *p2pSyncState, list *sobject.SharedObjectList) {
	// Replace the listed addresses and restart on a change.
	udp := &state.udp
	udp.mtx.Lock()
	defer udp.mtx.Unlock()
	udp.listed = a.listedUDPPeers(list)
	if err := a.restartStaticUDPLocked(state); err != nil && state.ctx.Err() == nil {
		a.le.WithError(err).Warn("failed to start static udp transport")
	}
}

// dialJoiningOwner adds an invitation owner's static address to the running
// generation, so the join and the sync that follows share one link. The
// address stays dialed for the generation even if the join fails.
func (a *ProviderAccount) dialJoiningOwner(owner peer.ID, addr string) error {
	// Resolve the running generation.
	var state *p2pSyncState
	a.p2pSyncBcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		state = a.p2pSync
	})
	if state == nil {
		return errors.New("P2P sync is not running")
	}

	// Add the owner and restart on a change.
	udp := &state.udp
	udp.mtx.Lock()
	defer udp.mtx.Unlock()
	if udp.joining == nil {
		udp.joining = make(map[peer.ID]string, 1)
	}
	udp.joining[owner] = addr
	return a.restartStaticUDPLocked(state)
}

// restartStaticUDPLocked restarts the transport when the dialed set changed.
func (a *ProviderAccount) restartStaticUDPLocked(state *p2pSyncState) error {
	// Keep the running transport while the set is unchanged.
	udp := &state.udp
	var peers map[peer.ID]string
	if len(udp.joining) != 0 || len(udp.listed) != 0 {
		peers = maps.Clone(udp.joining)
		if peers == nil {
			peers = make(map[peer.ID]string, len(udp.listed))
		}
		maps.Copy(peers, udp.listed)
	}
	if maps.Equal(peers, udp.peers) {
		return nil
	}

	// Stop the previous transport, then dial the new set.
	if udp.cancel != nil {
		udp.cancel()
		udp.peers, udp.cancel = nil, nil
	}
	if len(peers) == 0 {
		return nil
	}
	ctx, cancel := context.WithCancel(state.ctx)
	if _, err := state.sessionTransport.StartUDP(ctx, transport.UDPConfig{Peers: peers, NamedOnly: true}); err != nil {
		cancel()
		return err
	}
	udp.peers, udp.cancel = peers, cancel
	return nil
}
