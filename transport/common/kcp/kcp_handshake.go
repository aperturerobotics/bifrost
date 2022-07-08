package kcp

import (
	"context"
	"net"
	"sync"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/aperturerobotics/bifrost/peer"
	"google.golang.org/protobuf/proto"
)

// inflightHandshake is an on-going handshake.
type inflightHandshake struct {
	ctxCancel context.CancelFunc
	addr      net.Addr

	mtx         sync.Mutex
	completeErr error
	complete    bool
	lnk         *Link
	hs          identity.Handshaker
	pendingData [][]byte
	completeCbs []func(err error)
}

func (h *inflightHandshake) pushPacket(packet []byte) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if h.lnk != nil {
		h.lnk.handleRawPacket(packet)
		return
	}

	if h.complete {
		b2 := make([]byte, len(packet))
		copy(b2, packet)
		h.pendingData = append(h.pendingData, b2)
	} else {
		h.complete = !h.hs.Handle(packet)
	}
}

func (h *inflightHandshake) pushCompleteCb(cb func(err error)) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if h.complete {
		go cb(h.completeErr)
		return
	}

	h.completeCbs = append(h.completeCbs, cb)
}

// handleCompleteHandshake handles a completed handshake.
func (u *Transport) handleCompleteHandshake(
	addr net.Addr,
	result *identity.Result,
	initiator bool,
	extraData [][]byte,
) *Link {
	ctx := u.ctx
	as := addr.String()
	pid, _ := peer.IDFromPublicKey(result.Peer)

	le := u.le.
		WithField("remote-id", pid.Pretty()).
		WithField("remote-addr", as)

	exd := &HandshakeExtraData{}
	if err := proto.Unmarshal(result.ExtraData, exd); err != nil {
		le.WithError(err).Warn("unable to decode extra data from handshake")
		exd.Reset()
	} else {
		le = le.WithField("remote-transport-id", exd.GetLocalTransportUuid())
	}
	le.Debug("handshake complete")

	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	if l, ok := u.links[as]; ok {
		le.
			Debug("userping old session with peer")
		u.linksMtx.Unlock()
		l.Close()
		u.linksMtx.Lock()
	}

	var lnk *Link
	var err error
	lnk, err = NewLink(
		ctx,
		le,
		u.opts,
		u.peerID,
		u.pc.LocalAddr(),
		addr,
		u.GetUUID(),
		exd.GetLocalTransportUuid(),
		result,
		func(b []byte, a net.Addr) (int, error) {
			/*
				le.
					WithField("data-len", len(b)).
					WithField("addr", a.String()).
					Debug("writing packet")
			*/
			return u.pc.WriteTo(b, a)
		},
		initiator,
		func() {
			if lnk != nil {
				go u.handleLinkLost(as, lnk)
			}
		},
	)
	if err != nil {
		le.WithError(err).Warn("cannot construct link, dropping conn")
		return nil
	}

	u.links[as] = lnk
	for _, b := range extraData {
		lnk.handleRawPacket(b)
	}
	go u.handler.HandleLinkEstablished(lnk)
	return lnk
}

// pushHandshaker builds a new handshaker for the address.
// it is expected that handshakesMtx is locked before calling pushHandshaker
func (u *Transport) pushHandshaker(
	ctx context.Context,
	addr net.Addr,
	inititiator bool,
) (*inflightHandshake, error) {
	as := addr.String()
	nctx, nctxCancel := context.WithTimeout(ctx, handshakeTimeout)
	hs := &inflightHandshake{ctxCancel: nctxCancel, addr: addr}
	var err error
	// TODO: construct extra data
	ed := &HandshakeExtraData{LocalTransportUuid: u.GetUUID()}
	edDat, err := proto.Marshal(ed)
	if err != nil {
		return nil, err
	}
	hs.hs, err = s2s.NewHandshaker(
		u.privKey,
		nil,
		func(data []byte) error {
			data = append(data, byte(PacketType_PacketType_HANDSHAKE))
			/*
				u.le.
					WithField("data-len", len(data)).
					WithField("addr", addr.String()).
					Debugf("writing handshaking packet: %v", data)
			*/
			_, err := u.pc.WriteTo(data, addr)
			return err
		},
		nil,
		inititiator,
		edDat,
	)
	if err != nil {
		nctxCancel()
		return nil, err
	}

	if old, ok := u.handshakes[as]; ok && old.ctxCancel != nil {
		old.ctxCancel()
	}

	u.handshakes[as] = hs
	go u.processHandshake(nctx, hs, inititiator)
	return hs, nil
}

// processHandshake processes an in-flight handshake.
func (u *Transport) processHandshake(
	ctx context.Context,
	hs *inflightHandshake,
	initiator bool,
) {
	as := hs.addr.String()
	ule := u.le.WithField("addr", as)

	defer func() {
		hs.hs.Close()
		u.handshakesMtx.Lock()
		ohs := u.handshakes[as]
		if ohs == hs {
			delete(u.handshakes, as)
		}
		u.handshakesMtx.Unlock()
	}()

	res, err := hs.hs.Execute(ctx)
	if err != nil && err != context.Canceled {
		ule.WithError(err).Warn("error handshaking")
	}

	hs.mtx.Lock()
	if err == nil {
		hs.lnk = u.handleCompleteHandshake(hs.addr, res, initiator, hs.pendingData)
	}
	hs.complete = true
	hs.completeErr = err
	hs.pendingData = nil
	for _, cb := range hs.completeCbs {
		go cb(err)
	}
	hs.completeCbs = nil
	hs.mtx.Unlock()
}
